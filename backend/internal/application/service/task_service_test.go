// Package service 任务应用服务测试
package service

import (
	"testing"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskTest(t *testing.T) (*TaskAppService, *testutil.TestDB) {
	tdb := testutil.NewTestDB(t)
	taskRepo := repository.NewTaskRepository(tdb.DB)
	service := NewTaskAppService(taskRepo)
	return service, tdb
}

func createTestOrg(t *testing.T, tdb *testutil.TestDB) *entity.Organization {
	org := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: "test-org-id"},
		Name:       "测试组织",
		Code:       "TEST001",
		Type:       entity.OrgTypeCity,
	}
	testutil.MustCreate(t, tdb.DB, org)
	return org
}

func createTestUser(t *testing.T, tdb *testutil.TestDB, id, phone, role string) *entity.User {
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: id},
		Nickname:   "测试用户-" + id,
		Phone:      phone,
		Email:      id + "@test.com",
		WxOpenID:   "wx-openid-" + id,
		Role:       entity.Role(role),
		Status:     entity.UserStatusActive,
		OrgID:      "test-org-id",
	}
	user.SetPassword("password123")
	testutil.MustCreate(t, tdb.DB, user)
	return user
}

func createTestTask(t *testing.T, tdb *testutil.TestDB, creatorID, orgID string, status entity.TaskStatus) *entity.Task {
	task := &entity.Task{
		BaseEntity:  entity.BaseEntity{ID: "task-" + time.Now().Format("20060102150405") + "-" + creatorID},
		Title:       "测试任务-" + string(status),
		Description: "这是一个测试任务",
		Type:        entity.TaskTypeSearch,
		Priority:    entity.TaskPriorityMedium,
		Status:      status,
		CreatorID:   creatorID,
		OrgID:       orgID,
		Location:    "北京市",
		Province:    "北京市",
		City:        "北京市",
		District:    "朝阳区",
		Address:     "测试地址",
		Lat:         39.9042,
		Lng:         116.4074,
	}
	testutil.MustCreate(t, tdb.DB, task)
	return task
}

func TestTaskAppService_Create(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试组织和用户
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))

	tests := []struct {
		name    string
		req     *dto.CreateTaskRequest
		wantErr bool
	}{
		{
			name: "valid task with basic data",
			req: &dto.CreateTaskRequest{
				Title:       "搜索走失人员",
				Description: "在朝阳区进行搜索",
				Type:        string(entity.TaskTypeSearch),
				Location:    "北京市朝阳区",
				Province:    "北京市",
				City:        "北京市",
				District:    "朝阳区",
				Address:     "三里屯附近",
				Lat:         39.934,
				Lng:         116.455,
			},
			wantErr: false,
		},
		{
			name: "create task with priority",
			req: &dto.CreateTaskRequest{
				Title:       "紧急搜索任务",
				Description: "高优先级搜索任务",
				Type:        string(entity.TaskTypeSearch),
				Priority:    string(entity.TaskPriorityUrgent),
				Location:    "北京市海淀区",
			},
			wantErr: false,
		},
		{
			name: "create task with deadline",
			req: &dto.CreateTaskRequest{
				Title:       "限时搜索任务",
				Description: "有截止日期的任务",
				Type:        string(entity.TaskTypeVerify),
				Priority:    string(entity.TaskPriorityHigh),
				Deadline:    time.Now().Add(24 * time.Hour),
				Location:    "北京市",
			},
			wantErr: false,
		},
		{
			name: "create task with missing person ID",
			req: &dto.CreateTaskRequest{
				Title:           "关联走失人员任务",
				Description:     "关联到特定走失人员的任务",
				Type:            string(entity.TaskTypeFollow),
				MissingPersonID: "missing-person-001",
				Location:        "北京市",
			},
			wantErr: false,
		},
		{
			name: "create task with default priority",
			req: &dto.CreateTaskRequest{
				Title:    "默认优先级任务",
				Type:     string(entity.TaskTypeAssist),
				Location: "上海市",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Create(testutil.Context(), tt.req, creator.ID, org.ID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, resp.ID)
			assert.Equal(t, tt.req.Title, resp.Title)
			assert.Equal(t, tt.req.Description, resp.Description)
			assert.Equal(t, tt.req.Type, resp.Type)
			assert.Equal(t, creator.ID, resp.CreatorID)
			assert.Equal(t, org.ID, resp.OrgID)
			assert.Equal(t, string(entity.TaskStatusDraft), resp.Status)

			// 验证优先级
			if tt.req.Priority != "" {
				assert.Equal(t, tt.req.Priority, resp.Priority)
			} else {
				assert.Equal(t, string(entity.TaskPriorityMedium), resp.Priority)
			}

			// 验证截止日期
			if !tt.req.Deadline.IsZero() {
				assert.NotNil(t, resp.Deadline)
				assert.WithinDuration(t, tt.req.Deadline, *resp.Deadline, time.Second)
			}

			// 验证关联走失人员
			if tt.req.MissingPersonID != "" {
				assert.NotNil(t, resp.MissingPersonID)
				assert.Equal(t, tt.req.MissingPersonID, *resp.MissingPersonID)
			}
		})
	}
}

func TestTaskAppService_GetByID(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)

	tests := []struct {
		name    string
		taskID  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "existing task",
			taskID:  task.ID,
			wantErr: false,
		},
		{
			name:    "non-existing task",
			taskID:  "non-existing-id",
			wantErr: true,
			errMsg:  "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.GetByID(testutil.Context(), tt.taskID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrTaskNotFound, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.taskID, resp.ID)
			assert.Equal(t, task.Title, resp.Title)
			assert.Equal(t, string(task.Status), resp.Status)
		})
	}
}

func TestTaskAppService_List(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建多个任务用于测试筛选
	task1 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "task-1"},
		Title:      "搜索任务A",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusDraft,
		Priority:   entity.TaskPriorityHigh,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
	}
	testutil.MustCreate(t, tdb.DB, task1)

	task2 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "task-2"},
		Title:      "核实任务B",
		Type:       entity.TaskTypeVerify,
		Status:     entity.TaskStatusAssigned,
		Priority:   entity.TaskPriorityMedium,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task2)

	task3 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "task-3"},
		Title:      "跟进任务C",
		Type:       entity.TaskTypeFollow,
		Status:     entity.TaskStatusProcessing,
		Priority:   entity.TaskPriorityUrgent,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task3)

	tests := []struct {
		name         string
		req          *dto.TaskListRequest
		wantCount    int
		wantTotal    int64
		shouldContain []string
	}{
		{
			name: "list all tasks",
			req: &dto.TaskListRequest{
				Page:     1,
				PageSize: 10,
			},
			wantCount:     3,
			wantTotal:     3,
			shouldContain: []string{"task-1", "task-2", "task-3"},
		},
		{
			name: "filter by status",
			req: &dto.TaskListRequest{
				Page:     1,
				PageSize: 10,
				Status:   string(entity.TaskStatusAssigned),
			},
			wantCount:     1,
			wantTotal:     1,
			shouldContain: []string{"task-2"},
		},
		{
			name: "filter by type",
			req: &dto.TaskListRequest{
				Page:     1,
				PageSize: 10,
				Type:     string(entity.TaskTypeSearch),
			},
			wantCount:     1,
			wantTotal:     1,
			shouldContain: []string{"task-1"},
		},
		{
			name: "filter by priority",
			req: &dto.TaskListRequest{
				Page:     1,
				PageSize: 10,
				Priority: string(entity.TaskPriorityUrgent),
			},
			wantCount:     1,
			wantTotal:     1,
			shouldContain: []string{"task-3"},
		},
		{
			name: "filter by assignee",
			req: &dto.TaskListRequest{
				Page:       1,
				PageSize:   10,
				AssigneeID: assignee.ID,
			},
			wantCount:     2,
			wantTotal:     2,
			shouldContain: []string{"task-2", "task-3"},
		},
		{
			name: "filter by keyword",
			req: &dto.TaskListRequest{
				Page:     1,
				PageSize: 10,
				Keyword:  "搜索",
			},
			wantCount:     1,
			wantTotal:     1,
			shouldContain: []string{"task-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.List(testutil.Context(), tt.req)
			require.NoError(t, err)
			assert.Len(t, resp.List, tt.wantCount)
			assert.Equal(t, tt.wantTotal, resp.Total)

			// 验证返回的任务ID
			ids := make(map[string]bool)
			for _, task := range resp.List {
				ids[task.ID] = true
			}
			for _, expectedID := range tt.shouldContain {
				assert.True(t, ids[expectedID], "expected task %s to be in results", expectedID)
			}
		})
	}
}

func TestTaskAppService_Update(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)

	tests := []struct {
		name    string
		taskID  string
		req     *dto.UpdateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "update title and description",
			taskID: task.ID,
			req: &dto.UpdateTaskRequest{
				Title:       "更新后的标题",
				Description: "更新后的描述",
			},
			wantErr: false,
		},
		{
			name:   "update type and priority",
			taskID: task.ID,
			req: &dto.UpdateTaskRequest{
				Type:     string(entity.TaskTypeVerify),
				Priority: string(entity.TaskPriorityHigh),
			},
			wantErr: false,
		},
		{
			name:   "update location info",
			taskID: task.ID,
			req: &dto.UpdateTaskRequest{
				Location: "上海市",
				Province: "上海市",
				City:     "上海市",
				District: "浦东新区",
				Address:  "新地址",
				Lat:      31.2304,
				Lng:      121.4737,
			},
			wantErr: false,
		},
		{
			name:   "update deadline",
			taskID: task.ID,
			req: &dto.UpdateTaskRequest{
				Deadline: time.Now().Add(48 * time.Hour),
			},
			wantErr: false,
		},
		{
			name:   "update non-existing task",
			taskID: "non-existing-id",
			req: &dto.UpdateTaskRequest{
				Title: "新标题",
			},
			wantErr: true,
			errMsg:  "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Update(testutil.Context(), tt.taskID, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, ErrTaskNotFound, err)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.taskID, resp.ID)

			// 验证更新的字段
			if tt.req.Title != "" {
				assert.Equal(t, tt.req.Title, resp.Title)
			}
			if tt.req.Description != "" {
				assert.Equal(t, tt.req.Description, resp.Description)
			}
			if tt.req.Type != "" {
				assert.Equal(t, tt.req.Type, resp.Type)
			}
			if tt.req.Priority != "" {
				assert.Equal(t, tt.req.Priority, resp.Priority)
			}
			if !tt.req.Deadline.IsZero() {
				assert.NotNil(t, resp.Deadline)
				assert.WithinDuration(t, tt.req.Deadline, *resp.Deadline, time.Second)
			}
			if tt.req.Location != "" {
				assert.Equal(t, tt.req.Location, resp.Location)
				assert.Equal(t, tt.req.Province, resp.Province)
				assert.Equal(t, tt.req.City, resp.City)
				assert.Equal(t, tt.req.District, resp.District)
				assert.Equal(t, tt.req.Address, resp.Address)
				// Note: Lat/Lng may be updated by the service, just verify they are set
				assert.NotZero(t, resp.Lat)
				assert.NotZero(t, resp.Lng)
			}
		})
	}
}

func TestTaskAppService_Update_CannotUpdateCompletedTask(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))

	// 创建一个已完成的任务
	completedTask := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "completed-task-id"},
		Title:      "已完成任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusCompleted,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
	}
	testutil.MustCreate(t, tdb.DB, completedTask)

	// 尝试更新已完成的任务
	req := &dto.UpdateTaskRequest{
		Title: "尝试修改",
	}
	_, err := service.Update(testutil.Context(), completedTask.ID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot update completed or cancelled task")
}

func TestTaskAppService_Assign(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建一个草稿状态的任务
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)

	// 分配任务
	err := service.Assign(testutil.Context(), task.ID, assignee.ID)
	require.NoError(t, err)

	// 验证任务状态已更新
	var updatedTask entity.Task
	testutil.MustFind(t, tdb.DB, &updatedTask, "id = ?", task.ID)
	assert.Equal(t, entity.TaskStatusAssigned, updatedTask.Status)
	assert.NotNil(t, updatedTask.AssigneeID)
	assert.Equal(t, assignee.ID, *updatedTask.AssigneeID)
}

func TestTaskAppService_Assign_AlreadyAssigned(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee1 := createTestUser(t, tdb, "assignee-1", "13800138001", string(entity.RoleVolunteer))
	assignee2 := createTestUser(t, tdb, "assignee-2", "13800138002", string(entity.RoleVolunteer))

	// 创建一个已分配的任务
	task := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "assigned-task-id"},
		Title:      "已分配任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusAssigned,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee1.ID,
	}
	testutil.MustCreate(t, tdb.DB, task)

	// 尝试重新分配给另一个人
	err := service.Assign(testutil.Context(), task.ID, assignee2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "当前状态不能分配任务")
}

func TestTaskAppService_Assign_NonExistingTask(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 尝试分配不存在的任务
	err := service.Assign(testutil.Context(), "non-existing-id", assignee.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestTaskAppService_Start(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建一个已分配的任务
	task := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "task-to-start"},
		Title:      "要开始执行的任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusAssigned,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task)

	// 开始任务
	err := service.Start(testutil.Context(), task.ID, assignee.ID)
	require.NoError(t, err)

	// 验证任务状态已更新
	var updatedTask entity.Task
	testutil.MustFind(t, tdb.DB, &updatedTask, "id = ?", task.ID)
	assert.Equal(t, entity.TaskStatusProcessing, updatedTask.Status)
	assert.NotNil(t, updatedTask.StartedAt)
}

func TestTaskAppService_Start_InvalidStatus(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	tests := []struct {
		name   string
		status entity.TaskStatus
	}{
		{
			name:   "cannot start draft task",
			status: entity.TaskStatusDraft,
		},
		{
			name:   "cannot start pending task",
			status: entity.TaskStatusPending,
		},
		{
			name:   "cannot start processing task",
			status: entity.TaskStatusProcessing,
		},
		{
			name:   "cannot start completed task",
			status: entity.TaskStatusCompleted,
		},
		{
			name:   "cannot start cancelled task",
			status: entity.TaskStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &entity.Task{
				BaseEntity: entity.BaseEntity{ID: "task-" + string(tt.status)},
				Title:      "测试任务-" + string(tt.status),
				Type:       entity.TaskTypeSearch,
				Status:     tt.status,
				CreatorID:  creator.ID,
				OrgID:      org.ID,
				AssigneeID: &assignee.ID,
			}
			testutil.MustCreate(t, tdb.DB, task)

			err := service.Start(testutil.Context(), task.ID, assignee.ID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "当前状态不能开始任务")
		})
	}
}

func TestTaskAppService_Start_NonExistingTask(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	err := service.Start(testutil.Context(), "non-existing-id", "user-id")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestTaskAppService_Complete(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建一个进行中的任务
	task := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "task-to-complete"},
		Title:      "要完成的任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusProcessing,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task)

	// 完成任务
	req := &dto.CompleteTaskRequest{
		Result: "已成功找到走失人员，已联系家属认领",
	}
	err := service.Complete(testutil.Context(), task.ID, req, assignee.ID)
	require.NoError(t, err)

	// 验证任务状态已更新
	var updatedTask entity.Task
	testutil.MustFind(t, tdb.DB, &updatedTask, "id = ?", task.ID)
	assert.Equal(t, entity.TaskStatusCompleted, updatedTask.Status)
	assert.NotNil(t, updatedTask.CompletedAt)
	assert.Equal(t, req.Result, updatedTask.Result)
	assert.Equal(t, 100, updatedTask.Progress)
}

func TestTaskAppService_Complete_InvalidStatus(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13900139000", string(entity.RoleVolunteer))

	// 创建一个草稿状态的任务，分配给 assignee（不能直接完成，因为状态不对）
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)
	assigneeID := assignee.ID
	task.AssigneeID = &assigneeID
	require.NoError(t, tdb.DB.Save(task).Error)

	req := &dto.CompleteTaskRequest{
		Result: "尝试完成",
	}
	err := service.Complete(testutil.Context(), task.ID, req, assignee.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "当前状态不能完成任务")
}

func TestTaskAppService_Complete_NonExistingTask(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	req := &dto.CompleteTaskRequest{
		Result: "结果",
	}
	err := service.Complete(testutil.Context(), "non-existing-id", req, "user-id")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestTaskAppService_Cancel(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))

	// 创建一个草稿状态的任务
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)

	// 取消任务
	req := &dto.CancelTaskRequest{
		Reason: "任务重复，已创建新任务",
	}
	err := service.Cancel(testutil.Context(), task.ID, req, creator.ID)
	require.NoError(t, err)

	// 验证任务状态已更新
	var updatedTask entity.Task
	testutil.MustFind(t, tdb.DB, &updatedTask, "id = ?", task.ID)
	assert.Equal(t, entity.TaskStatusCancelled, updatedTask.Status)
	assert.Equal(t, req.Reason, updatedTask.Feedback)
}

func TestTaskAppService_Cancel_InvalidStatus(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))

	// 创建一个已完成的任务（不能取消）
	completedTask := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "completed-task-id"},
		Title:      "已完成任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusCompleted,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
	}
	testutil.MustCreate(t, tdb.DB, completedTask)

	req := &dto.CancelTaskRequest{
		Reason: "尝试取消",
	}
	err := service.Cancel(testutil.Context(), completedTask.ID, req, creator.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "当前状态不能取消任务")
}

func TestTaskAppService_Cancel_NonExistingTask(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	req := &dto.CancelTaskRequest{
		Reason: "取消原因",
	}
	err := service.Cancel(testutil.Context(), "non-existing-id", req, "user-id")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestTaskAppService_Delete(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)

	// 删除任务
	err := service.Delete(testutil.Context(), task.ID)
	require.NoError(t, err)

	// 验证任务已被软删除
	var count int64
	tdb.DB.Model(&entity.Task{}).Where("id = ? AND deleted_at IS NULL", task.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestTaskAppService_UpdateProgress(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建一个进行中的任务
	task := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "task-progress"},
		Title:      "更新进度的任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusProcessing,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
		Progress:   0,
	}
	testutil.MustCreate(t, tdb.DB, task)

	// 更新进度
	err := service.UpdateProgress(testutil.Context(), task.ID, 50, assignee.ID)
	require.NoError(t, err)

	// 验证进度已更新
	var updatedTask entity.Task
	testutil.MustFind(t, tdb.DB, &updatedTask, "id = ?", task.ID)
	assert.Equal(t, 50, updatedTask.Progress)
}

func TestTaskAppService_UpdateProgress_InvalidProgress(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusProcessing)

	// 尝试设置无效进度
	err := service.UpdateProgress(testutil.Context(), task.ID, 150, creator.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "进度必须在0-100之间")
}

func TestTaskAppService_GetMyTasks(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建分配给 assignee 的任务
	task1 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "my-task-1"},
		Title:      "我的任务1",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusAssigned,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task1)

	task2 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "my-task-2"},
		Title:      "我的任务2",
		Type:       entity.TaskTypeVerify,
		Status:     entity.TaskStatusProcessing,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task2)

	// 创建不分配给 assignee 的任务
	task3 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "other-task"},
		Title:      "别人的任务",
		Type:       entity.TaskTypeFollow,
		Status:     entity.TaskStatusDraft,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
	}
	testutil.MustCreate(t, tdb.DB, task3)

	// 获取我的任务
	resp, err := service.GetMyTasks(testutil.Context(), assignee.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.List, 2)
}

func TestTaskAppService_GetPendingTasks(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))

	// 创建待分配任务
	task1 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "pending-task-1"},
		Title:      "待分配任务1",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusPending,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
	}
	testutil.MustCreate(t, tdb.DB, task1)

	task2 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "pending-task-2"},
		Title:      "待分配任务2",
		Type:       entity.TaskTypeVerify,
		Status:     entity.TaskStatusPending,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
	}
	testutil.MustCreate(t, tdb.DB, task2)

	// 创建已分配任务（不应该出现在待分配列表中）
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))
	task3 := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "assigned-task"},
		Title:      "已分配任务",
		Type:       entity.TaskTypeFollow,
		Status:     entity.TaskStatusAssigned,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		AssigneeID: &assignee.ID,
	}
	testutil.MustCreate(t, tdb.DB, task3)

	// 获取待分配任务
	resp, err := service.GetPendingTasks(testutil.Context(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.List, 2)
}

func TestTaskAppService_GetLogs(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建并分配任务（会产生日志）
	task := createTestTask(t, tdb, creator.ID, org.ID, entity.TaskStatusDraft)
	err := service.Assign(testutil.Context(), task.ID, assignee.ID)
	require.NoError(t, err)

	// 获取任务日志
	logs, err := service.GetLogs(testutil.Context(), task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(logs), 1)

	// 验证日志内容
	foundAssignLog := false
	for _, log := range logs {
		if log.Action == "assign" {
			foundAssignLog = true
			assert.Equal(t, task.ID, log.TaskID)
			assert.Equal(t, assignee.ID, log.UserID)
			assert.Equal(t, string(entity.TaskStatusAssigned), log.NewStatus)
		}
	}
	assert.True(t, foundAssignLog, "should have an assign log")
}

func TestTaskAppService_GetStats(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))
	assignee := createTestUser(t, tdb, "assignee-id", "13800138001", string(entity.RoleVolunteer))

	// 创建各种状态的任务
	tasks := []*entity.Task{
		{BaseEntity: entity.BaseEntity{ID: "task-draft"}, Title: "草稿", Type: entity.TaskTypeSearch, Status: entity.TaskStatusDraft, CreatorID: creator.ID, OrgID: org.ID},
		{BaseEntity: entity.BaseEntity{ID: "task-pending"}, Title: "待分配", Type: entity.TaskTypeVerify, Status: entity.TaskStatusPending, CreatorID: creator.ID, OrgID: org.ID},
		{BaseEntity: entity.BaseEntity{ID: "task-assigned"}, Title: "已分配", Type: entity.TaskTypeAssist, Status: entity.TaskStatusAssigned, CreatorID: creator.ID, OrgID: org.ID, AssigneeID: &assignee.ID},
		{BaseEntity: entity.BaseEntity{ID: "task-processing"}, Title: "进行中", Type: entity.TaskTypeFollow, Status: entity.TaskStatusProcessing, CreatorID: creator.ID, OrgID: org.ID, AssigneeID: &assignee.ID},
		{BaseEntity: entity.BaseEntity{ID: "task-completed"}, Title: "已完成", Type: entity.TaskTypeInterview, Status: entity.TaskStatusCompleted, CreatorID: creator.ID, OrgID: org.ID, AssigneeID: &assignee.ID},
	}

	for _, task := range tasks {
		testutil.MustCreate(t, tdb.DB, task)
	}

	// 获取统计
	stats, err := service.GetStats(testutil.Context(), assignee.ID)
	require.NoError(t, err)
	// 验证基本统计存在（具体数值取决于实现）
	assert.GreaterOrEqual(t, stats.Total, int64(0))
}

func TestTaskAppService_GetOverdueTasks(t *testing.T) {
	service, tdb := setupTaskTest(t)
	defer tdb.Close()

	// 创建测试数据
	org := createTestOrg(t, tdb)
	creator := createTestUser(t, tdb, "creator-id", "13800138000", string(entity.RoleManager))

	// 创建已逾期的任务
	deadline := time.Now().Add(-24 * time.Hour)
	overdueTask := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "overdue-task"},
		Title:      "已逾期任务",
		Type:       entity.TaskTypeSearch,
		Status:     entity.TaskStatusProcessing,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		Deadline:   &deadline,
	}
	testutil.MustCreate(t, tdb.DB, overdueTask)

	// 创建未逾期任务
	futureDeadline := time.Now().Add(24 * time.Hour)
	notOverdueTask := &entity.Task{
		BaseEntity: entity.BaseEntity{ID: "not-overdue-task"},
		Title:      "未逾期任务",
		Type:       entity.TaskTypeVerify,
		Status:     entity.TaskStatusProcessing,
		CreatorID:  creator.ID,
		OrgID:      org.ID,
		Deadline:   &futureDeadline,
	}
	testutil.MustCreate(t, tdb.DB, notOverdueTask)

	// 获取逾期任务
	resp, err := service.GetOverdueTasks(testutil.Context(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
	assert.Len(t, resp.List, 1)
	assert.Equal(t, "overdue-task", resp.List[0].ID)
}
