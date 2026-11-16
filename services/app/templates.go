// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"html/template"
	"io/fs"
	"strings"
	"svnova/models"
)

// ParseTemplates uses Gitea's slash-separated template names independently of leaf filenames.
func ParseTemplates(source fs.FS) (*template.Template, error) {
	t := template.New("").Funcs(template.FuncMap{"historyActions": historyActions, "historyAction": historyAction, "historySummary": historySummary, "fileSize": fileSize, "repositorySize": repositorySize, "buildVersion": buildVersion, "orgLink": orgLink, "teamRepositoriesLink": teamRepositoriesLink, "teamLink": teamLink, "repoLink": repoLink, "orgSection": orgSection})
	err := fs.WalkDir(source, "templates", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(name, ".tmpl") {
			return nil
		}
		content, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		key := strings.TrimSuffix(strings.TrimPrefix(name, "templates/"), ".tmpl")
		if _, err = t.New(key).Parse(string(content)); err != nil {
			return fmt.Errorf("parse template %s: %w", key, err)
		}
		return nil
	})
	return t, err
}

func fileSize(size int64) string {
	if size < 0 {
		return "—"
	}
	return humanize.IBytes(uint64(size))
}
func repositorySize(repo models.Repository) string {
	if repo.SizeUpdated == 0 {
		return "待统计"
	}
	return fileSize(repo.Size)
}
