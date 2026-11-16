// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image/color"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"svnova/models"
	"svnova/modules/avatar/identicon"
	"sync"
)

func currentAvatar(w http.ResponseWriter, u *models.User) {
	avatarImageSize(w, u.Name, 512)
}

type avatarCacheKey struct {
	Seed string
	Size int
}

var generatedAvatars = struct {
	sync.Mutex
	images map[avatarCacheKey][]byte
}{images: make(map[avatarCacheKey][]byte)}

func avatarSize(r *http.Request, fallback int) int {
	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err == nil {
		switch size {
		case 48, 56, 64, 80, 96, 128, 200, 256, 512:
			return size
		}
	}
	return fallback
}
func avatarPNG(seed string, size int) []byte {
	key := avatarCacheKey{seed, size}
	generatedAvatars.Lock()
	defer generatedAvatars.Unlock()
	if data, ok := generatedAvatars.images[key]; ok {
		return data
	}
	var out bytes.Buffer
	_ = png.Encode(&out, identicon.New(size, color.White, identicon.DarkColors).Make([]byte(seed)))
	// Bound memory even when many distinct accounts are viewed.
	if len(generatedAvatars.images) >= 256 {
		clear(generatedAvatars.images)
	}
	data := out.Bytes()
	generatedAvatars.images[key] = data
	return data
}
func avatarImageSize(w http.ResponseWriter, seed string, size int) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(avatarPNG(seed, size))
}
func serveAvatar(w http.ResponseWriter, r *http.Request, seed string, size int) {
	data := avatarPNG(seed, size)
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))
	w.Header().Set("Content-Type", "image/png")
	// Revalidate every request so revoked access cannot reuse a fresh browser response.
	w.Header().Set("Cache-Control", "private, no-cache")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if r.Method != "HEAD" {
		_, _ = w.Write(data)
	}
}

func (a *App) organizationAvatar(w http.ResponseWriter, r *http.Request, u *models.User) {
	orgID := id(strings.TrimPrefix(r.URL.Path, "/avatars/organizations/"))
	org := new(models.Organization)
	ok, err := a.DB.ID(orgID).Get(org)
	if err != nil {
		http.Error(w, "头像加载失败", 500)
		return
	}
	if !ok || orgID <= 0 {
		http.NotFound(w, r)
		return
	}
	if !u.Admin {
		ok, err = a.DB.Where("org_id=? AND user_id=?", orgID, u.ID).Exist(new(models.Membership))
		if err != nil {
			http.Error(w, "头像加载失败", 500)
			return
		}
		if !ok {
			http.Error(w, "没有访问权限", 403)
			return
		}
	}
	serveAvatar(w, r, "organization:"+org.Name, avatarSize(r, 64))
}

func (a *App) memberAvatar(w http.ResponseWriter, r *http.Request, u *models.User) {
	userID := id(strings.TrimPrefix(r.URL.Path, "/avatars/users/"))
	orgID := id(r.URL.Query().Get("org"))
	target := new(models.User)
	ok, err := a.DB.ID(userID).Get(target)
	if err != nil {
		http.Error(w, "头像加载失败", 500)
		return
	}
	if !ok || userID <= 0 {
		http.NotFound(w, r)
		return
	}
	if !u.Admin && userID != u.ID {
		viewer, err := a.DB.Where("org_id=? AND user_id=?", orgID, u.ID).Exist(new(models.Membership))
		if err != nil || !viewer {
			http.Error(w, "没有访问权限", 403)
			return
		}
		member, err := a.DB.Where("org_id=? AND user_id=?", orgID, userID).Exist(new(models.Membership))
		if err != nil || !member {
			http.Error(w, "没有访问权限", 403)
			return
		}
	}
	serveAvatar(w, r, target.Name, avatarSize(r, 64))
}
