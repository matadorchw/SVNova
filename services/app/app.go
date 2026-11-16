// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"svnova/models"
	"svnova/services/directory"
	"svnova/services/notifications"
	"svnova/services/security"
	"svnova/services/storage"
	"svnova/services/svn"
	"svnova/services/tasks"
	"xorm.io/xorm"
)

type App struct {
	Notifications *notifications.Service
	DB            *xorm.Engine
	Vault         *security.Vault
	SVN           svn.Runner
	Worker        *tasks.Worker
	Directory     *directory.Directory
	Data          string
	WorkRoot      string
	Config        Config
	Templates     *template.Template
	Assets        http.Handler
	mu            sync.RWMutex
	installMu     sync.Mutex
	attempts      map[string]attempt
}
type attempt struct {
	Count int
	Since time.Time
}

func New(data, bin string, t *template.Template, assets http.Handler) (*App, error) {
	if err := PrepareWorkPath(data); err != nil {
		return nil, err
	}
	a := &App{WorkRoot: data, Data: filepath.Join(data, "data"), SVN: svn.Runner{Bin: bin}, Templates: t, Assets: assets, attempts: map[string]attempt{}}
	installed, err := a.loadConfig()
	if err != nil {
		return nil, err
	}
	if !installed {
		return a, nil
	}
	if _, err = os.Stat(a.Config.SecretKey); err != nil {
		return nil, errors.New("installed application encryption key missing; restore the original secret.key")
	}
	err = a.open()
	if err != nil && a.DB != nil {
		a.DB.Close()
		a.DB = nil
	}
	return a, err
}
func (a *App) open() error {
	var err error
	a.DB, err = models.Open(a.Config.Database)
	if err != nil {
		return err
	}
	a.Vault, err = security.NewVault(a.Config.SecretKey)
	if err != nil {
		return err
	}
	if err = storage.CheckParents(a.Config.Repositories); err != nil {
		return err
	}
	if err = os.MkdirAll(a.Config.Repositories, 0700); err != nil {
		return err
	}
	if err = a.recoverConversions(); err != nil {
		return err
	}
	if err = storage.Relocate(context.Background(), a.DB, a.Config.Repositories, a.SVN); err != nil {
		return err
	}
	a.Worker = &tasks.Worker{DB: a.DB, SVN: a.SVN, Vault: a.Vault, Root: a.Config.Repositories, Convert: a.executeConversion}
	a.Directory = &directory.Directory{DB: a.DB, Vault: a.Vault}
	a.Notifications = &notifications.Service{DB: a.DB, Vault: a.Vault, Origin: a.Config.Origin}
	if err = a.Notifications.Recover(); err != nil {
		return err
	}
	if err = a.Directory.RecoverSyncRuns(); err != nil {
		return err
	}
	return nil
}
func (a *App) Background(ctx context.Context) {
	var jobs sync.WaitGroup
	defer jobs.Wait()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	started := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.RLock()
			ready := a.DB != nil
			w := a.Worker
			d := a.Directory
			notifier := a.Notifications
			a.mu.RUnlock()
			if !ready {
				continue
			}
			if !started {
				if d != nil {
					jobs.Add(1)
					go func() { defer jobs.Done(); d.RunScheduler(ctx) }()
				}
				if notifier != nil {
					jobs.Add(1)
					go func() { defer jobs.Done(); notifier.Run(ctx) }()
				}
				jobs.Add(1)
				go func() { defer jobs.Done(); w.Run(ctx) }()
				started = true
			}
		}
	}
}
func (a *App) Handler() http.Handler { return http.HandlerFunc(a.serve) }
func (a *App) page(w http.ResponseWriter, name string, data any) {
	switch name {
	case "login":
		name = "user/auth/signin"
	case "app":
		name = "base/page"
	}
	switch value := data.(type) {
	case map[string]any:
		value["SiteName"] = a.siteName()
	case *webPage:
		value.SiteName = a.siteName()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template %s failed: %v", name, err)
	}
}
func (a *App) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		a.Assets.ServeHTTP(w, r)
		return
	}
	a.mu.RLock()
	ready := a.DB != nil
	a.mu.RUnlock()
	if !ready {
		a.install(w, r)
		return
	}
	if r.URL.Path == "/install" {
		http.Error(w, "安装已锁定", http.StatusForbidden)
		return
	}
	if r.URL.Path == "/login" && (r.Method == "GET" || r.Method == "HEAD") {
		http.Redirect(w, r, "/user/login", 303)
		return
	}
	if r.URL.Path == "/login" || r.URL.Path == "/user/login" {
		a.login(w, r)
		return
	}
	u, s := a.current(r)
	if u == nil {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, "请登录", 401)
		} else {
			http.Redirect(w, r, "/user/login", 303)
		}
		return
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		if r.Header.Get("X-CSRF-Token") != s.CSRF || !a.sameOrigin(r) {
			http.Error(w, "请求验证失败", 403)
			return
		}
	}
	if strings.HasPrefix(r.URL.Path, "/avatars/users/") && (r.Method == "GET" || r.Method == "HEAD") {
		a.memberAvatar(w, r, u)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/avatars/organizations/") && (r.Method == "GET" || r.Method == "HEAD") {
		a.organizationAvatar(w, r, u)
		return
	}
	if r.URL.Path == "/avatars/current" && (r.Method == "GET" || r.Method == "HEAD") {
		serveAvatar(w, r, u.Name, avatarSize(r, 512))
		return
	}
	if (r.URL.Path == "/logout" || r.URL.Path == "/user/logout") && r.Method == "POST" {
		_, _ = a.DB.ID(s.Token).Delete(new(models.Session))
		http.SetCookie(w, &http.Cookie{Name: "svnova_session", Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
		http.Redirect(w, r, "/user/login", 303)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin") && !u.Admin {
		http.Error(w, "仅管理员可访问", 403)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		if err := a.api(w, r, u); err != nil {
			var e *httpError
			var commandError *svn.CommandError
			if errors.As(err, &e) {
				http.Error(w, e.Message, e.Status)
			} else if errors.As(err, &commandError) {
				http.Error(w, commandError.UserMessage(), http.StatusBadRequest)
			} else {
				log.Printf("request failed: %v", err)
				http.Error(w, "操作失败，请检查输入或任务状态", 400)
			}
		}
		return
	}
	if a.migrationAction(w, r, u) {
		return
	}
	if a.redirectLegacy(w, r, u) {
		return
	}
	a.web(w, r, u, s.CSRF)
}
func (a *App) sameOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" {
		return false
	}
	u, e := url.Parse(o)
	if e != nil {
		return false
	}
	if a.Config.Origin != "" {
		expected, _ := url.Parse(a.Config.Origin)
		return expected != nil && u.Scheme == expected.Scheme && u.Host == expected.Host
	}
	return u.Host == r.Host
}
func (a *App) current(r *http.Request) (*models.User, *models.Session) {
	c, err := r.Cookie("svnova_session")
	if err != nil {
		return nil, nil
	}
	sum := sha256.Sum256([]byte(c.Value))
	s := new(models.Session)
	ok, err := a.DB.ID(hex.EncodeToString(sum[:])).Get(s)
	if err != nil || !ok || s.Expires < time.Now().Unix() {
		return nil, nil
	}
	u := new(models.User)
	ok, err = a.DB.ID(s.UserID).Get(u)
	if err != nil || !ok || !u.Active {
		return nil, nil
	}
	if u.LDAPAccount {
		if _, err := a.Directory.Source(); err != nil {
			return nil, nil
		}
	}
	return u, s
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	_, ldapErr := a.Directory.Source()
	if r.Method == "GET" {
		a.page(w, "login", map[string]any{"LDAP": ldapErr == nil})
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	if !a.sameOrigin(r) {
		http.Error(w, "请求验证失败", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if r.ParseForm() != nil {
		http.Error(w, "无效输入", 400)
		return
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	a.mu.Lock()
	v := a.attempts[ip]
	if time.Since(v.Since) > 15*time.Minute {
		v = attempt{Since: time.Now()}
	}
	v.Count++
	if len(a.attempts) > 10000 {
		for k, x := range a.attempts {
			if time.Since(x.Since) > 15*time.Minute {
				delete(a.attempts, k)
			}
		}
	}
	a.attempts[ip] = v
	a.mu.Unlock()
	if v.Count > 20 {
		http.Error(w, "尝试过于频繁，请稍后重试", 429)
		return
	}
	var user *models.User
	var err error
	if r.FormValue("source") == "ldap" {
		user, err = a.Directory.Login(r.FormValue("name"), r.FormValue("password"))
	} else {
		u := new(models.User)
		ok, e := a.DB.Where("name=? AND ldap_account=0", r.FormValue("name")).Get(u)
		err = e
		if ok && u.Active && security.Verify(r.FormValue("password"), u.Password, u.Salt) {
			user = u
		}
	}
	if err != nil || user == nil {
		a.page(w, "login", map[string]any{"Error": "登录失败，请检查账号、密码及认证源", "LDAP": ldapErr == nil})
		return
	}
	raw := security.Token()
	sum := sha256.Sum256([]byte(raw))
	s := &models.Session{Token: hex.EncodeToString(sum[:]), UserID: user.ID, CSRF: security.Token(), Expires: time.Now().Add(12 * time.Hour).Unix()}
	if err = a.createSession(user, s); err != nil {
		http.Error(w, "无法建立会话", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "svnova_session", Value: raw, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(a.Config.Origin, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: 43200})
	http.Redirect(w, r, "/", 303)
}
func (a *App) install(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/install" {
		http.Redirect(w, r, "/install", 303)
		return
	}
	if r.Method == "GET" {
		a.page(w, "install", map[string]any{"Data": a.Data})
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	a.installMu.Lock()
	defer a.installMu.Unlock()
	a.mu.RLock()
	ready := a.DB != nil
	a.mu.RUnlock()
	if ready {
		http.Error(w, "安装已锁定", 403)
		return
	}
	origin := r.Header.Get("Origin")
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host != r.Host {
		http.Error(w, "请求验证失败", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if r.ParseForm() != nil {
		http.Error(w, "无效输入", 400)
		return
	}
	name, password := r.FormValue("name"), r.FormValue("password")
	if !validName(name) || len(password) < 12 {
		a.page(w, "install", map[string]any{"Data": a.Data, "Error": "用户名格式无效，或密码少于12个字符"})
		return
	}
	siteName := "SVNova"
	if r.Form.Has("site_name") {
		siteName, err = normalizeSiteName(r.FormValue("site_name"))
		if err != nil {
			a.page(w, "install", map[string]any{"Data": a.Data, "Error": err.Error()})
			return
		}
	}
	cfg := Config{Listen: r.FormValue("listen"), Origin: strings.TrimRight(r.FormValue("origin"), "/"), Repositories: r.FormValue("repositories"), Database: r.FormValue("database")}
	if cfg.Listen == "" {
		cfg.Listen = "127.0.0.1:3000"
	}
	if !filepath.IsAbs(cfg.Repositories) || !filepath.IsAbs(cfg.Database) {
		http.Error(w, "请使用绝对路径", 400)
		return
	}
	u, e := url.Parse(cfg.Origin)
	if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		http.Error(w, "访问地址无效", 400)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg.SecretKey = filepath.Join(a.Data, "secret.key")
	a.Config = cfg
	if err = a.open(); err != nil {
		if a.DB != nil {
			a.DB.Close()
		}
		a.DB = nil
		http.Error(w, "无法初始化数据目录", 500)
		return
	}
	h, salt, err := security.Password(password)
	if err == nil {
		_, err = a.DB.Insert(&models.User{Name: name, Password: h, Salt: salt, Admin: true, Active: true})
	}
	if err == nil {
		_, err = a.DB.Insert(&models.SiteSetting{ID: 1, SiteName: siteName})
	}
	if err == nil {
		err = a.saveConfig()
	}
	if err != nil {
		a.DB.Close()
		a.DB = nil
		http.Error(w, "安装未完成，请检查数据库及配置", 500)
		return
	}
	http.Redirect(w, r, "/user/login", 303)
}

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,79}$`)

func validName(s string) bool { return namePattern.MatchString(s) && s != "." && s != ".." }

type httpError struct {
	Status  int
	Message string
}

func (e *httpError) Error() string { return e.Message }
func forbidden() error             { return &httpError{403, "没有访问权限"} }
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func send(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	return json.NewEncoder(w).Encode(v)
}
func id(s string) int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }
