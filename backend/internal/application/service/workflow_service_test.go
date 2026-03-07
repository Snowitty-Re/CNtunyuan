package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	pkgErrors "github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkflowRepository 工作流仓储Mock
type MockWorkflowRepository struct {
	mock.Mock
}

func (m *MockWorkflowRepository) Create(ctx context.Context, workflow *entity.WorkflowDefinition) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) FindByID(ctx context.Context, id string) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowRepository) FindByKey(ctx context.Context, key string, version int) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, key, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowRepository) Update(ctx context.Context, workflow *entity.WorkflowDefinition) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkflowRepository) List(ctx context.Context, query *entity.WorkflowQuery) ([]*entity.WorkflowDefinition, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*entity.WorkflowDefinition), args.Get(1).(int64), args.Error(2)
}

func (m *MockWorkflowRepository) Publish(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockWorkflowTaskRepository 工作流任务仓储Mock
type MockWorkflowTaskRepository struct {
	mock.Mock
}

func (m *MockWorkflowTaskRepository) Create(ctx context.Context, task *entity.WorkflowTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockWorkflowTaskRepository) FindByID(ctx context.Context, id string) (*entity.WorkflowTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowTask), args.Error(1)
}

func (m *MockWorkflowTaskRepository) FindByInstanceID(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error) {
	args := m.Called(ctx, instanceID)
	return args.Get(0).([]*entity.WorkflowTask), args.Error(1)
}

func (m *MockWorkflowTaskRepository) FindPendingByAssignee(ctx context.Context, assigneeID string, page, pageSize int) ([]*entity.WorkflowTask, int64, error) {
	args := m.Called(ctx, assigneeID, page, pageSize)
	return args.Get(0).([]*entity.WorkflowTask), args.Get(1).(int64), args.Error(2)
}

func (m *MockWorkflowTaskRepository) Update(ctx context.Context, task *entity.WorkflowTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockWorkflowTaskRepository) Complete(ctx context.Context, taskID string, action entity.WorkflowAction, comment string, userID string) error {
	args := m.Called(ctx, taskID, action, comment, userID)
	return args.Error(0)
}

// MockUserRepositoryForWorkflow 用户仓储Mock
type MockUserRepositoryForWorkflow struct {
	mock.Mock
}

func (m *MockUserRepositoryForWorkflow) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestWorkflowAppService_CreateDefinition(t *testing.T) {
	workflowRepo := new(MockWorkflowRepository)
	taskRepo := new(MockWorkflowTaskRepository)
	userRepo := new(MockUserRepositoryForWorkflow)
	service := NewWorkflowAppService(workflowRepo, taskRepo, userRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.CreateWorkflowDefinitionRequest
		mock    func()
		wantErr bool
		errCode pkgErrors.ErrorCode
	}{
		{
			name: "success - create workflow definition",
			req: &dto.CreateWorkflowDefinitionRequest{
				Name:        "请假审批",
				Key:         "leave_request",
				Description: "员工请假审批流程",
				Category:    "人事",
				OrgID:       "org-001",
			},
			mock: func() {
				workflowRepo.On("Create", ctx, mock.AnythingOfType("*entity.WorkflowDefinition")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "fail - duplicate key",
			req: &dto.CreateWorkflowDefinitionRequest{
				Name:     "请假审批",
				Key:      "leave_request",
				Category: "人事",
				OrgID:    "org-001",
			},
			mock: func() {
				workflowRepo.On("Create", ctx, mock.AnythingOfType("*entity.WorkflowDefinition")).Return(errors.New("duplicate key")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeWorkflowInstanceExists,
		},
		{
			name: "fail - empty name",
			req: &dto.CreateWorkflowDefinitionRequest{
				Name:     "",
				Key:      "leave_request",
				Category: "人事",
				OrgID:    "org-001",
			},
			mock:    func() {},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.CreateDefinition(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			workflowRepo.AssertExpectations(t)
		})
	}
}

func TestWorkflowAppService_GetDefinition(t *testing.T) {
	workflowRepo := new(MockWorkflowRepository)
	taskRepo := new(MockWorkflowTaskRepository)
	userRepo := new(MockUserRepositoryForWorkflow)
	service := NewWorkflowAppService(workflowRepo, taskRepo, userRepo)
	ctx := context.Background()

	tests := []struct {
		name         string
		workflowID   string
		mock         func()
		wantErr      bool
		errCode      pkgErrors.ErrorCode
	}{
		{
			name:       "success - get workflow definition",
			workflowID: "wf-001",
			mock: func() {
				workflow := &entity.WorkflowDefinition{
					ID:          "wf-001",
					Name:        "请假审批",
					Key:         "leave_request",
					Description: "员工请假审批流程",
					Status:      entity.WorkflowStatusActive,
					Version:     1,
					CreatedAt:   time.Now(),
				}
				workflowRepo.On("FindByID", ctx, "wf-001").Return(workflow, nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "fail - workflow not found",
			workflowID: "wf-notfound",
			mock: func() {
				workflowRepo.On("FindByID", ctx, "wf-notfound").Return(nil, errors.New("not found")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeWorkflowNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.GetDefinition(ctx, tt.workflowID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.workflowID, resp.ID)
			}

			workflowRepo.AssertExpectations(t)
		})
	}
}

func TestWorkflowAppService_ListDefinitions(t *testing.T) {
	workflowRepo := new(MockWorkflowRepository)
	taskRepo := new(MockWorkflowTaskRepository)
	userRepo := new(MockUserRepositoryForWorkflow)
	service := NewWorkflowAppService(workflowRepo, taskRepo, userRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.WorkflowListRequest
		mock    func()
		wantErr bool
	}{
		{
			name: "success - list workflow definitions",
			req: &dto.WorkflowListRequest{
				Page:     1,
				PageSize: 10,
				Keyword:  "请假",
				Category: "人事",
				Status:   "active",
			},
			mock: func() {
				workflows := []*entity.WorkflowDefinition{
					{ID: "wf-001", Name: "请假审批", Key: "leave_request"},
					{ID: "wf-002", Name: "加班申请", Key: "overtime_request"},
				}
				workflowRepo.On("List", ctx, mock.AnythingOfType("*entity.WorkflowQuery")).Return(workflows, int64(2), nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.ListDefinitions(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.GreaterOrEqual(t, len(resp.List), 0)
			}

			workflowRepo.AssertExpectations(t)
		})
	}
}

func TestWorkflowAppService_ApproveTask(t *testing.T) {
	workflowRepo := new(MockWorkflowRepository)
	taskRepo := new(MockWorkflowTaskRepository)
	userRepo := new(MockUserRepositoryForWorkflow)
	service := NewWorkflowAppService(workflowRepo, taskRepo, userRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		taskID  string
		userID  string
		req     *dto.ApproveTaskRequest
		mock    func()
		wantErr bool
		errCode pkgErrors.ErrorCode
	}{
		{
			name:   "success - approve task",
			taskID: "task-001",
			userID: "user-001",
			req: &dto.ApproveTaskRequest{
				Action:  "approve",
				Comment: "同意",
			},
			mock: func() {
				task := &entity.WorkflowTask{
					ID:         "task-001",
					AssigneeID: "user-001",
					Status:     entity.WorkflowTaskStatusPending,
				}
				taskRepo.On("FindByID", ctx, "task-001").Return(task, nil).Once()
				taskRepo.On("Complete", ctx, "task-001", entity.WorkflowActionApprove, "同意", "user-001").Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "fail - task not found",
			taskID: "task-notfound",
			userID: "user-001",
			req: &dto.ApproveTaskRequest{
				Action: "approve",
			},
			mock: func() {
				taskRepo.On("FindByID", ctx, "task-notfound").Return(nil, errors.New("not found")).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeWorkflowNotFound,
		},
		{
			name:   "fail - invalid action",
			taskID: "task-001",
			userID: "user-001",
			req: &dto.ApproveTaskRequest{
				Action: "invalid",
			},
			mock: func() {
				task := &entity.WorkflowTask{
					ID:         "task-001",
					AssigneeID: "user-001",
					Status:     entity.WorkflowTaskStatusPending,
				}
				taskRepo.On("FindByID", ctx, "task-001").Return(task, nil).Once()
			},
			wantErr: true,
			errCode: pkgErrors.CodeInvalidParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := service.ApproveTask(ctx, tt.taskID, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, pkgErrors.GetCode(err))
				}
			} else {
				assert.NoError(t, err)
			}

			workflowRepo.AssertExpectations(t)
			taskRepo.AssertExpectations(t)
		})
	}
}

func TestWorkflowAppService_GetTodoList(t *testing.T) {
	workflowRepo := new(MockWorkflowRepository)
	taskRepo := new(MockWorkflowTaskRepository)
	userRepo := new(MockUserRepositoryForWorkflow)
	service := NewWorkflowAppService(workflowRepo, taskRepo, userRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  string
		req     *dto.WorkflowTodoRequest
		mock    func()
		wantErr bool
	}{
		{
			name:   "success - get todo list",
			userID: "user-001",
			req: &dto.WorkflowTodoRequest{
				Page:     1,
				PageSize: 10,
			},
			mock: func() {
				tasks := []*entity.WorkflowTask{
					{ID: "task-001", AssigneeID: "user-001", Status: entity.WorkflowTaskStatusPending},
					{ID: "task-002", AssigneeID: "user-001", Status: entity.WorkflowTaskStatusPending},
				}
				taskRepo.On("FindPendingByAssignee", ctx, "user-001", 1, 10).Return(tasks, int64(2), nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			resp, err := service.GetTodoList(ctx, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			taskRepo.AssertExpectations(t)
		})
	}
}
