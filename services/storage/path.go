// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var component = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,79}$`)

func ValidateName(name string) error {
	if !component.MatchString(name) || strings.HasSuffix(name, ".") {
		return errors.New("name is not a Windows directory component")
	}
	stem := strings.ToUpper(strings.Split(name, ".")[0])
	if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || (len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9') {
		return errors.New("reserved Windows device name")
	}
	return nil
}
func RepositoryPath(root, org, name string) (string, error) {
	if err := ValidateName(org); err != nil {
		return "", fmt.Errorf("organization %q: %w", org, err)
	}
	if err := ValidateName(name); err != nil {
		return "", fmt.Errorf("repository %q: %w", name, err)
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := filepath.Join(absolute, org, name)
	if !Within(absolute, target) {
		return "", errors.New("repository outside root")
	}
	return target, nil
}
func StagingPath(root string, id int64) string {
	return filepath.Join(root, ".svnova-staging", strconv.FormatInt(id, 10))
}
func Within(root, target string) bool {
	root, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// Missing leaves are allowed, but every existing ancestor must be a real directory.
func CheckParents(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for current := absolute; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil {
			if !info.IsDir() || linked(current, info) {
				return fmt.Errorf("linked or non-directory path: %s", current)
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if filepath.Dir(current) == current {
			break
		}
	}
	return nil
}
func RejectLinks(root string) error {
	if err := CheckParents(root); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if linked(path, info) {
			return errors.New("linked repository paths are not supported")
		}
		return nil
	})
}
