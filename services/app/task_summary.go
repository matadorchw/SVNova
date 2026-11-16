// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"svnova/models"
)

func executionSummary(task models.Task, execution *models.TaskExecution) string {
	switch task.State {
	case "failed":
		return (&webPage{}).TaskError(task.Error)
	case "interrupted":
		return "执行已中断"
	case "succeeded":
	default:
		return "—"
	}
	known := execution != nil && execution.State == "succeeded" && execution.RevisionKnown && execution.BeforeRevision >= 0 && execution.AfterRevision >= execution.BeforeRevision
	switch task.Kind {
	case "sync":
		if !known {
			return "同步完成"
		}
		if execution.AfterRevision == execution.BeforeRevision {
			return fmt.Sprintf("无新增修订（r%d）", execution.AfterRevision)
		}
		return fmt.Sprintf("新增 %d 个修订（r%d → r%d）", execution.AfterRevision-execution.BeforeRevision, execution.BeforeRevision, execution.AfterRevision)
	case "conversion":
		if known {
			return fmt.Sprintf("已转换为镜像（r%d），未逐修订比对历史内容", execution.AfterRevision)
		}
		return "转换完成，未逐修订比对历史内容"
	default:
		if known {
			return fmt.Sprintf("已迁移至 r%d", execution.AfterRevision)
		}
		return "迁移完成"
	}
}

// Load only the latest attempt for each visible task, without per-row queries.
func (a *App) latestTaskExecutions(ids []int64) (map[int64]*models.TaskExecution, error) {
	out := map[int64]*models.TaskExecution{}
	if len(ids) == 0 {
		return out, nil
	}
	var executions []models.TaskExecution
	if err := a.DB.In("task_id", ids).Where("id=(SELECT MAX(latest.id) FROM task_execution latest WHERE latest.task_id=task_execution.task_id)").Find(&executions); err != nil {
		return nil, err
	}
	for i := range executions {
		out[executions[i].TaskID] = &executions[i]
	}
	return out, nil
}
