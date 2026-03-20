// Package entity 任务实体测试
package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		taskType  TaskType
		creatorID string
		orgID     string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid task",
			title:     "测试任务",
			taskType:  TaskTypeSearch,
			creatorID: "creator-id",
			orgID:     "org-id",
			wantErr:   false,
		},
		{
			name:      "empty title",
			title:     "",
			taskType:  TaskTypeSearch,
			creatorID: "creator-id",
			orgID:     "org-id",
			wantErr:   true,
			errMsg:    "任务标题不能为空",
		},
		{
			name:      "empty type",
			title:     "测试任务",
			taskType:  "",
			creatorID: "creator-id",
			orgID:     "org-id",
			wantErr:   true,
			errMsg:    "任务类型不能为空",
		},
		{
			name:      "invalid type",
			title:     "测试任务",
			taskType:  "invalid_type",
			creatorID: "creator-id",
			orgID:     "org-id",
			wantErr:   true,
			errMsg:    "无效的任务类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.title, tt.taskType, tt.creatorID, tt.orgID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, task.ID)
			assert.Equal(t, tt.title, task.Title)
			assert.Equal(t, tt.taskType, task.Type)
			assert.Equal(t, TaskStatusDraft, task.Status)
			assert.Equal(t, TaskPriorityMedium, task.Priority)
		})
	}
}

func TestTask_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"draft", TaskStatusDraft, true},
		{"pending", TaskStatusPending, true},
		{"assigned", TaskStatusAssigned, true},
		{"processing", TaskStatusProcessing, true},
		{"completed", TaskStatusCompleted, false},
		{"cancelled", TaskStatusCancelled, false},
		{"overdue", TaskStatusOverdue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			assert.Equal(t, tt.want, task.IsActive())
		})
	}
}

func TestTask_CanAssign(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"draft", TaskStatusDraft, true},
		{"pending", TaskStatusPending, true},
		{"assigned", TaskStatusAssigned, false},
		{"processing", TaskStatusProcessing, false},
		{"completed", TaskStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			assert.Equal(t, tt.want, task.CanAssign())
		})
	}
}

func TestTask_CanStart(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"draft", TaskStatusDraft, false},
		{"pending", TaskStatusPending, false},
		{"assigned", TaskStatusAssigned, true},
		{"processing", TaskStatusProcessing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			if tt.status == TaskStatusAssigned {
				assigneeID := "test-assignee"
				task.AssigneeID = &assigneeID
			}
			assert.Equal(t, tt.want, task.CanStart())
		})
	}
}

func TestTask_CanComplete(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"assigned", TaskStatusAssigned, true},
		{"processing", TaskStatusProcessing, true},
		{"draft", TaskStatusDraft, false},
		{"completed", TaskStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			assert.Equal(t, tt.want, task.CanComplete())
		})
	}
}

func TestTask_Assign(t *testing.T) {
	tests := []struct {
		name       string
		status     TaskStatus
		assigneeID string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "assign from draft",
			status:     TaskStatusDraft,
			assigneeID: "user-1",
			wantErr:    false,
		},
		{
			name:       "assign from pending",
			status:     TaskStatusPending,
			assigneeID: "user-1",
			wantErr:    false,
		},
		{
			name:       "cannot assign from processing",
			status:     TaskStatusProcessing,
			assigneeID: "user-1",
			wantErr:    true,
			errMsg:     "当前状态不能分配任务",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			err := task.Assign(tt.assigneeID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, &tt.assigneeID, task.AssigneeID)
			assert.Equal(t, TaskStatusAssigned, task.Status)
		})
	}
}

func TestTask_Start(t *testing.T) {
	tests := []struct {
		name    string
		status  TaskStatus
		wantErr bool
		errMsg  string
	}{
		{
			name:    "start from assigned",
			status:  TaskStatusAssigned,
			wantErr: false,
		},
		{
			name:    "cannot start from draft",
			status:  TaskStatusDraft,
			wantErr: true,
			errMsg:  "当前状态不能开始任务",
		},
		{
			name:    "cannot start from processing",
			status:  TaskStatusProcessing,
			wantErr: true,
			errMsg:  "当前状态不能开始任务",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			if tt.status == TaskStatusAssigned {
				assigneeID := "test-assignee"
				task.AssigneeID = &assigneeID
			}
			err := task.Start()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, TaskStatusProcessing, task.Status)
			assert.NotNil(t, task.StartedAt)
		})
	}
}

func TestTask_Complete(t *testing.T) {
	tests := []struct {
		name    string
		status  TaskStatus
		wantErr bool
		errMsg  string
	}{
		{
			name:    "complete from processing",
			status:  TaskStatusProcessing,
			wantErr: false,
		},
		{
			name:    "complete from assigned",
			status:  TaskStatusAssigned,
			wantErr: false,
		},
		{
			name:    "cannot complete from draft",
			status:  TaskStatusDraft,
			wantErr: true,
			errMsg:  "当前状态不能完成任务",
		},
		{
			name:    "cannot complete from completed",
			status:  TaskStatusCompleted,
			wantErr: true,
			errMsg:  "当前状态不能完成任务",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			result := "任务已完成"
			err := task.Complete(result)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, TaskStatusCompleted, task.Status)
			assert.Equal(t, result, task.Result)
			assert.Equal(t, 100, task.Progress)
			assert.NotNil(t, task.CompletedAt)
		})
	}
}

func TestTask_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		status  TaskStatus
		wantErr bool
		errMsg  string
	}{
		{
			name:    "cancel from draft",
			status:  TaskStatusDraft,
			wantErr: false,
		},
		{
			name:    "cancel from processing",
			status:  TaskStatusProcessing,
			wantErr: false,
		},
		{
			name:    "cannot cancel from completed",
			status:  TaskStatusCompleted,
			wantErr: true,
			errMsg:  "当前状态不能取消任务",
		},
		{
			name:    "cannot cancel from cancelled",
			status:  TaskStatusCancelled,
			wantErr: true,
			errMsg:  "当前状态不能取消任务",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			reason := "取消原因"
			err := task.Cancel(reason)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, TaskStatusCancelled, task.Status)
			assert.Equal(t, reason, task.Feedback)
		})
	}
}

func TestTask_UpdateProgress(t *testing.T) {
	tests := []struct {
		name     string
		progress int
		wantErr  bool
		errMsg   string
	}{
		{"valid 0", 0, false, ""},
		{"valid 50", 50, false, ""},
		{"valid 100", 100, false, ""},
		{"invalid negative", -1, true, "进度必须在0-100之间"},
		{"invalid over 100", 101, true, "进度必须在0-100之间"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{}
			err := task.UpdateProgress(tt.progress)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.progress, task.Progress)
		})
	}
}

func TestTask_IsOverdue(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name    string
		status  TaskStatus
		deadline *time.Time
		want    bool
	}{
		{
			name:     "overdue and processing",
			status:   TaskStatusProcessing,
			deadline: &past,
			want:     true,
		},
		{
			name:     "not overdue yet",
			status:   TaskStatusProcessing,
			deadline: &future,
			want:     false,
		},
		{
			name:     "completed task not overdue",
			status:   TaskStatusCompleted,
			deadline: &past,
			want:     false,
		},
		{
			name:     "cancelled task not overdue",
			status:   TaskStatusCancelled,
			deadline: &past,
			want:     false,
		},
		{
			name:     "no deadline",
			status:   TaskStatusProcessing,
			deadline: nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				Status:   tt.status,
				Deadline: tt.deadline,
			}
			assert.Equal(t, tt.want, task.IsOverdue())
		})
	}
}

func TestTask_CheckAndUpdateOverdue(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)

	task := &Task{
		Status:   TaskStatusProcessing,
		Deadline: &past,
	}
	
	updated := task.CheckAndUpdateOverdue()
	assert.True(t, updated)
	assert.Equal(t, TaskStatusOverdue, task.Status)
	
	// Already overdue, should not update again
	updated = task.CheckAndUpdateOverdue()
	assert.False(t, updated)
}
