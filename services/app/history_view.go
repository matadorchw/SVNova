// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"strings"
	"svnova/services/svn"
)

type revisionAction struct{ Label, Code string }

func historyAction(action string) revisionAction {
	switch action {
	case "A":
		return revisionAction{"新增", "A"}
	case "M":
		return revisionAction{"修改", "M"}
	case "D":
		return revisionAction{"删除", "D"}
	case "R":
		return revisionAction{"替换", "R"}
	default:
		return revisionAction{"未知动作", "?"}
	}
}

func historyActions(entry svn.LogEntry) []revisionAction {
	seen := map[string]bool{}
	for _, path := range entry.Paths {
		seen[path.Action] = true
	}
	actions := []revisionAction{}
	for _, action := range []string{"A", "M", "D", "R"} {
		if seen[action] {
			actions = append(actions, historyAction(action))
			delete(seen, action)
		}
	}
	if len(seen) > 0 {
		actions = append(actions, historyAction(""))
	}
	return actions
}

func historySummary(message string) string {
	return strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(message)
}
