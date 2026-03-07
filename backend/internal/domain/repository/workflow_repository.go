package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	// Definition operations
	CreateDefinition(ctx context.Context, def *entity.WorkflowDefinition) error
	UpdateDefinition(ctx context.Context, def *entity.WorkflowDefinition) error
	DeleteDefinition(ctx context.Context, id string) error
	FindDefinitionByID(ctx context.Context, id string) (*entity.WorkflowDefinition, error)
	FindDefinitionByKey(ctx context.Context, key string, version int) (*entity.WorkflowDefinition, error)
	FindActiveDefinitionByKey(ctx context.Context, key string) (*entity.WorkflowDefinition, error)
	ListDefinitions(ctx context.Context, orgID, category string, page, pageSize int) ([]*entity.WorkflowDefinition, int64, error)
	GetLatestVersion(ctx context.Context, key string) (int, error)
	
	// Node operations
	CreateNode(ctx context.Context, node *entity.WorkflowNode) error
	UpdateNode(ctx context.Context, node *entity.WorkflowNode) error
	DeleteNode(ctx context.Context, id string) error
	FindNodeByID(ctx context.Context, id string) (*entity.WorkflowNode, error)
	FindNodesByWorkflowID(ctx context.Context, workflowID string) ([]*entity.WorkflowNode, error)
	DeleteNodesByWorkflowID(ctx context.Context, workflowID string) error
	
	// Instance operations
	CreateInstance(ctx context.Context, instance *entity.WorkflowInstance) error
	UpdateInstance(ctx context.Context, instance *entity.WorkflowInstance) error
	DeleteInstance(ctx context.Context, id string) error
	FindInstanceByID(ctx context.Context, id string) (*entity.WorkflowInstance, error)
	FindInstanceByBusinessKey(ctx context.Context, businessKey string) (*entity.WorkflowInstance, error)
	ListInstances(ctx context.Context, query *entity.WorkflowQuery) ([]*entity.WorkflowInstance, int64, error)
	ListMyInstances(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowInstance, int64, error)
	ListMyTodoInstances(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowInstance, int64, error)
	ListMyDoneInstances(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowInstance, int64, error)
	
	// Transition operations
	CreateTransition(ctx context.Context, transition *entity.WorkflowTransition) error
	FindTransitionsByInstanceID(ctx context.Context, instanceID string) ([]*entity.WorkflowTransition, error)
	
	// Stats
	GetInstanceStats(ctx context.Context, orgID string) (*entity.WorkflowStats, error)
	GetUserTaskStats(ctx context.Context, userID string) (map[string]int64, error)
}

// WorkflowTaskRepository 工作流任务仓储接口
type WorkflowTaskRepository interface {
	Create(ctx context.Context, task *entity.WorkflowTask) error
	Update(ctx context.Context, task *entity.WorkflowTask) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.WorkflowTask, error)
	FindByInstanceID(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error)
	FindByNodeID(ctx context.Context, instanceID, nodeID string) ([]*entity.WorkflowTask, error)
	List(ctx context.Context, query *entity.WorkflowTaskQuery) ([]*entity.WorkflowTask, int64, error)
	
	// Specific queries
	FindTodoTasks(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowTask, int64, error)
	FindDoneTasks(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowTask, int64, error)
	FindTasksByInstance(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error)
	FindActiveTasksByInstance(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error)
	FindActiveTaskByInstanceAndNode(ctx context.Context, instanceID, nodeID string) (*entity.WorkflowTask, error)
	
	// Count queries
	CountTodoByUser(ctx context.Context, userID string) (int64, error)
	CountByInstance(ctx context.Context, instanceID string) (int64, error)
	CountCompletedByInstanceAndNode(ctx context.Context, instanceID, nodeID string) (int64, error)
	
	// Delegation
	CreateDelegation(ctx context.Context, delegation *entity.WorkflowDelegation) error
	UpdateDelegation(ctx context.Context, delegation *entity.WorkflowDelegation) error
	FindDelegationByID(ctx context.Context, id string) (*entity.WorkflowDelegation, error)
	FindActiveDelegation(ctx context.Context, taskID string) (*entity.WorkflowDelegation, error)
	
	// Reminder
	CreateReminder(ctx context.Context, reminder *entity.WorkflowReminder) error
	UpdateReminder(ctx context.Context, reminder *entity.WorkflowReminder) error
	FindReminderByTaskID(ctx context.Context, taskID string) (*entity.WorkflowReminder, error)
}
