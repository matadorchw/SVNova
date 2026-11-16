// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/directory"
	"svnova/services/security"
	"testing"
	"time"
)

func TestNotificationSearchFiltersAndPagination(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	a := &App{DB: db}
	for i := 0; i < 45; i++ {
		if _, err = db.Insert(&models.NotificationDelivery{EventKey: fmt.Sprintf("search-%d", i), Kind: "test", Status: "pending", Object: fmt.Sprintf("生产对象 %d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Insert(&models.NotificationDelivery{EventKey: "failed-search", Kind: "svn", Status: "failed", Object: "other", LastError: "E215004"}); err != nil {
		t.Fatal(err)
	}
	page, err := a.notificationData(httptest.NewRequest("GET", "/?tab=deliveries&q=生产&kind=test&status=pending&sort=oldest&page=2", nil))
	if err != nil || page.Total != 45 || len(page.Deliveries) != 20 || page.Deliveries[0].ID != 21 || page.Pages != 3 || page.Pagination.First == "" || page.Pagination.Last == "" {
		t.Fatalf("page: %+v %v", page, err)
	}
	link := page.Pagination.Next
	for _, part := range []string{"kind=test", "status=pending", "sort=oldest", "page=3", "q=%E7%94%9F%E4%BA%A7"} {
		if !strings.Contains(link, part) {
			t.Fatal(link)
		}
	}
	for _, query := range []string{"q=E215004", "q=%2346", "kind=svn&status=failed", "sort=invalid&page=999"} {
		result, e := a.notificationData(httptest.NewRequest("GET", "/?tab=deliveries&"+query, nil))
		if e != nil || len(result.Deliveries) == 0 {
			t.Fatalf("%s: %+v %v", query, result, e)
		}
		if query != "sort=invalid&page=999" && (result.Total != 1 || result.Deliveries[0].ID != 46) {
			t.Fatalf("incorrect search: %+v", result)
		}
	}
}

func TestNotificationAdminCSRFEncryptionAndPages(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	vault, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	if _, err = db.Insert(admin, reader, &models.Repository{ID: 1, Kind: "mirror", Name: "<script>alert(1)</script>"}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Vault: vault, Directory: &directory.Directory{DB: db, Vault: vault}, Templates: templates, Config: Config{Origin: "http://example.com"}}
	sum := sha256.Sum256([]byte("session-token"))
	if _, err = db.Insert(&models.Session{Token: hex.EncodeToString(sum[:]), UserID: 1, CSRF: "csrf", Expires: time.Now().Add(time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body, csrf string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://example.com"+path, strings.NewReader(body))
		r.AddCookie(&http.Cookie{Name: "svnova_session", Value: "session-token"})
		r.Header.Set("Origin", "http://example.com")
		r.Header.Set("X-CSRF-Token", csrf)
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, r)
		return w
	}
	if w := call("POST", "/api/v1/admin/notifications/test", "{}", ""); w.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
	if w := call("POST", "/api/v1/admin/ldap", "{}", ""); w.Code != 403 {
		t.Fatal("LDAP configuration accepted missing CSRF")
	}
	if w := call("POST", "/api/v1/admin/ldap/preview", "{}", ""); w.Code != 403 {
		t.Fatal("LDAP preview accepted missing CSRF")
	}
	if w := call("POST", "/api/v1/admin/ldap-schedule", `{"Time":"30 2 * * *"}`, ""); w.Code != 403 {
		t.Fatal("schedule accepted missing CSRF")
	}
	if w := call("POST", "/api/v1/admin/ldap-schedule", `{"Time":"30 2 * * *"}`, "csrf"); w.Code != 200 {
		t.Fatal("schedule save failed", w.Body.String())
	}
	body := `{"Enabled":false,"SVNEnabled":true,"LDAPEnabled":true,"ReplaceWebhook":true,"ReplaceSecret":true,"Webhook":"https://oapi.dingtalk.com/robot/send?access_token=private-token","Secret":"private-secret","RetentionDays":90}`
	if w := call("POST", "/api/v1/admin/notifications", body, "csrf"); w.Code != 200 {
		t.Fatal("save failed", w.Code, w.Body.String())
	}
	w := call("GET", "/api/v1/admin/notifications", "", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "private-") || strings.Contains(w.Body.String(), "ExcludedRepositories") {
		t.Fatal("secret leaked in config")
	}
	settingsPage := call("GET", "/-/admin/notifications", "", "")
	if settingsPage.Code != 200 || strings.Contains(settingsPage.Body.String(), "排除的镜像仓库") || strings.Contains(settingsPage.Body.String(), "notification-exclude") {
		t.Fatal("removed exclusion UI remains")
	}
	if n, _ := db.Count(new(models.NotificationDelivery)); n != 0 {
		t.Fatal("save sent notification")
	}
	if w = call("POST", "/api/v1/admin/notifications/test", "{}", "csrf"); w.Code != 200 {
		t.Fatal("explicit test not queued", w.Body.String())
	}
	if n, _ := db.Count(new(models.NotificationDelivery)); n != 1 {
		t.Fatal("test did not queue")
	}
	for _, route := range []string{"/-/admin/notifications", "/-/admin/notifications?tab=deliveries", "/-/admin/notifications?tab=deliveries&delivery=1"} {
		w = call("GET", route, "", "")
		if w.Code != 200 || strings.Contains(w.Body.String(), "页面加载失败") || strings.Contains(w.Body.String(), "private-secret") || strings.Contains(w.Body.String(), "private-token") || strings.Contains(w.Body.String(), "<script>alert") {
			t.Fatal("unsafe or failed page", route, w.Body.String())
		}
		w = httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), reader, "csrf")
		if w.Code != 403 {
			t.Fatal("non-admin page access")
		}
	}
	for _, path := range []string{"/api/v1/admin/notifications", "/api/v1/admin/notifications/deliveries", "/api/v1/admin/notifications/deliveries/1"} {
		if err = a.api(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil), reader); err == nil {
			t.Fatal("non-admin API allowed")
		}
	}
	w = call("GET", "/api/v1/admin/notifications/deliveries/1", "", "")
	var detail map[string]json.RawMessage
	if err = json.Unmarshal(w.Body.Bytes(), &detail); err != nil || w.Code != 200 {
		t.Fatal("detail missing", err)
	}
	if w = call("POST", "/api/v1/admin/notifications/deliveries/1/redeliver", "{}", "csrf"); w.Code != 409 {
		t.Fatal("pending delivery duplicated")
	}
}
