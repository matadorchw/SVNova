// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"
	"svnova/services/storage"
)

type Config struct {
	Listen, Origin, Database, Repositories, SecretKey string
	MirrorConversionTimeout                           time.Duration
}

// PrepareWorkPath also rejects linked ancestors before creating writable folders.
func PrepareWorkPath(root string) error {
	for _, name := range []string{"custom/conf", "data", "log"} {
		path := filepath.Join(root, name)
		if err := storage.CheckParents(path); err != nil {
			return err
		}
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) loadConfig() (bool, error) {
	path := filepath.Join(a.WorkRoot, "custom", "conf", "app.ini")
	f, err := ini.Load(path)
	if os.IsNotExist(err) {
		// Keep existing paths and the original key when opening a legacy instance.
		b, legacyErr := os.ReadFile(filepath.Join(a.WorkRoot, "config.json"))
		if os.IsNotExist(legacyErr) {
			return false, nil
		}
		if legacyErr != nil {
			return false, legacyErr
		}
		if err = json.Unmarshal(b, &a.Config); err != nil {
			return false, err
		}
		a.Config.SecretKey = filepath.Join(a.WorkRoot, "secret.key")
		if _, err = os.Stat(a.Config.SecretKey); err != nil {
			return false, err
		}
		if err = a.saveConfig(); err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if !f.Section("security").Key("INSTALL_LOCK").MustBool(false) {
		return false, errors.New("existing app.ini must have INSTALL_LOCK enabled")
	}
	if f.Section("database").Key("DB_TYPE").String() != "sqlite3" {
		return false, errors.New("only sqlite3 is supported")
	}
	server := f.Section("server")
	a.Data = resolveConfigPath(a.WorkRoot, server.Key("APP_DATA_PATH").MustString("data"))
	a.Config.Listen = net.JoinHostPort(server.Key("HTTP_ADDR").MustString("127.0.0.1"), server.Key("HTTP_PORT").MustString("3000"))
	a.Config.Origin = strings.TrimRight(server.Key("ROOT_URL").String(), "/")
	a.Config.Database = resolveConfigPath(a.WorkRoot, f.Section("database").Key("PATH").MustString("data/svnova.db"))
	a.Config.Repositories = resolveConfigPath(a.WorkRoot, f.Section("repository").Key("ROOT").MustString("data/repositories"))
	a.Config.MirrorConversionTimeout, err = time.ParseDuration(f.Section("repository").Key("MIRROR_CONVERSION_TIMEOUT").MustString("24h"))
	if err != nil || a.Config.MirrorConversionTimeout <= 0 {
		return false, errors.New("MIRROR_CONVERSION_TIMEOUT must be a positive duration")
	}
	uri, err := url.Parse(f.Section("security").Key("SECRET_KEY_URI").String())
	if err != nil || uri.Scheme != "file" || uri.Host != "" || uri.Path == "" {
		return false, errors.New("SECRET_KEY_URI must reference a local key file")
	}
	key := uri.Path
	if len(key) > 2 && key[0] == '/' && key[2] == ':' {
		key = key[1:]
	}
	a.Config.SecretKey = filepath.FromSlash(key)
	if !filepath.IsAbs(a.Config.SecretKey) {
		return false, errors.New("encryption key path must be absolute")
	}
	return true, nil
}

func resolveConfigPath(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}

func (a *App) saveConfig() error {
	a.Config.MirrorConversionTimeout = conversionTimeout(a.Config.MirrorConversionTimeout)
	host, port, err := net.SplitHostPort(a.Config.Listen)
	if err != nil {
		return err
	}
	path := filepath.Join(a.WorkRoot, "custom", "conf", "app.ini")
	f, err := ini.Load(path)
	if os.IsNotExist(err) {
		f = ini.Empty()
	} else if err != nil {
		return err
	}
	values := map[string]map[string]string{
		"server":     {"APP_DATA_PATH": a.Data, "HTTP_ADDR": host, "HTTP_PORT": port, "ROOT_URL": a.Config.Origin},
		"database":   {"DB_TYPE": "sqlite3", "PATH": a.Config.Database},
		"repository": {"ROOT": a.Config.Repositories, "MIRROR_CONVERSION_TIMEOUT": conversionTimeout(a.Config.MirrorConversionTimeout).String()},
		"security":   {"INSTALL_LOCK": "true", "SECRET_KEY_URI": keyURI(a.Config.SecretKey)},
		"log":        {"MODE": "console,file", "ROOT_PATH": filepath.Join(a.WorkRoot, "log")},
	}
	for section, pairs := range values {
		for key, value := range pairs {
			f.Section(section).Key(key).SetValue(value)
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".app-*.ini")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = f.WriteTo(tmp)
	}
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}

func keyURI(path string) string {
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return (&url.URL{Scheme: "file", Path: path}).String()
}

func conversionTimeout(value time.Duration) time.Duration {
	if value == 0 {
		return 24 * time.Hour
	}
	return value
}
