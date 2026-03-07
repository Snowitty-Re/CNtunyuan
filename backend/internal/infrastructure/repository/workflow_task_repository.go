package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// WorkflowTaskRepositoryImpl 工作流任务仓储实现
type WorkflowTaskRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowTaskRepository 创建工作流任务仓储
func NewWorkflowTaskRepository(db *gorm.DB) repository.WorkflowTaskRepository {
	return &WorkflowTaskRepositoryImpl{db: db}
}

// Create 创建任务
func (r *WorkflowTaskRepositoryImpl) Create(ctx context.Context, task *entity.WorkflowTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// Update 更新任务
func (r *WorkflowTaskRepositoryImpl) Update(ctx context.Context, task *entity.WorkflowTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除任务
func (r *WorkflowTaskRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkflowTask{}, "id = ?", id).Error
}

// FindByID 根据ID查找
func (r *WorkflowTaskRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.WorkflowTask, error) {
	var task entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Preload("Instance.Definition").
		First(&task, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &task, err
}

// FindByInstanceID 根据实例ID查找
func (r *WorkflowTaskRepositoryImpl) FindByInstanceID(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error) {
	var tasks []*entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Where("workflow_instance_id = ?", instanceID).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// FindByNodeID 根据节点ID查找
func (r *WorkflowTaskRepositoryImpl) FindByNodeID(ctx context.Context, instanceID, nodeID string) ([]*entity.WorkflowTask, error) {
	var tasks []*entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Where("workflow_instance_id = ? AND workflow_node_id = ?", instanceID, nodeID).
		Find(&tasks).Error
	return tasks, err
}

// List 列表查询
func (r *WorkflowTaskRepositoryImpl) List(ctx context.Context, query *entity.WorkflowTaskQuery) ([]*entity.WorkflowTask, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.WorkflowTask{})
	
	if query.InstanceID != "" {
		db = db.Where("workflow_instance_id = ?", query.InstanceID)
	}
	if query.AssigneeID != "" {
		db = db.Where("assignee_id = ?", query.AssigneeID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.WorkflowInstanceID != "" {
		db = db.Where("workflow_instance_id = ?", query.WorkflowInstanceID)
	}
	if query.IsOverdue {
		db = db.Where("due_time IS NOT NULL AND due_time < NOW()")
	}
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	
	var tasks []*entity.WorkflowTask
	err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error
	
	return tasks, total, err
}

// FindTodoTasks 查找待办任务
func (r *WorkflowTaskRepositoryImpl) FindTodoTasks(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowTask, int64, error) {
	query := &entity.WorkflowTaskQuery{
		AssigneeID: userID,
		Status:     entity.TaskStatusPending,
		Page:       page,
		PageSize:   pageSize,
	}
	return r.List(ctx, query)
}

// FindDoneTasks 查找已办任务
func (r *WorkflowTaskRepositoryImpl) FindDoneTasks(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowTask, int64, error) {
	query := &entity.WorkflowTaskQuery{
		AssigneeID: userID,
		Status:     entity.TaskStatusCompleted,
		Page:       page,
		PageSize:   pageSize,
	}
	return r.List(ctx, query)
}

// FindTasksByInstance 查找实例的所有任务
func (r *WorkflowTaskRepositoryImpl) FindTasksByInstance(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error) {
	return r.FindByInstanceID(ctx, instanceID)
}

// FindActiveTasksByInstance 查找实例的活跃任务
func (r *WorkflowTaskRepositoryImpl) FindActiveTasksByInstance(ctx context.Context, instanceID string) ([]*entity.WorkflowTask, error) {
	var tasks []*entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Where("workflow_instance_id = ?", instanceID).
		Where("status IN ?", []entity.TaskStatus{entity.TaskStatusPending, entity.TaskStatusProcessing}).
		Find(&tasks).Error
	return tasks, err
}

// FindActiveTaskByInstanceAndNode 查找实例和节点的活跃任务
func (r *WorkflowTaskRepositoryImpl) FindActiveTaskByInstanceAndNode(ctx context.Context, instanceID, nodeID string) (*entity.WorkflowTask, error) {
	var task entity.WorkflowTask
	err := r.db.WithContext(ctx).
		Where("workflow_instance_id = ? AND workflow_node_id = ?", instanceID, nodeID).
		Where("status IN ?", []entity.TaskStatus{entity.TaskStatusPending, entity.TaskStatusProcessing}).
		First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &task, err
}

// CountTodoByUser 统计用户待办数量
func (r *WorkflowTaskRepositoryImpl) CountTodoByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.WorkflowTask{}).
		Where("assignee_id = ?", userID).
		Where("status IN ?", []entity.TaskStatus{entity.TaskStatusPending, entity.TaskStatusProcessing}).
		Count(&count).Error
	return count, err
}

// CountByInstance 统计实例任务数量
func (r *WorkflowTaskRepositoryImpl) CountByInstance(ctx context.Context, instanceID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.WorkflowTask{}).
		Where("workflow_instance_id = ?", instanceID).
		Count(&count).Error
	return count, err
}

// CountCompletedByInstanceAndNode 统计实例节点已完成任务数
func (r *WorkflowTaskRepositoryImpl) CountCompletedByInstanceAndNode(ctx context.Context, instanceID, nodeID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.WorkflowTask{}).
		Where("workflow_instance_id = ? AND workflow_node_id = ?", instanceID, nodeID).
		Where("status = ?", entity.TaskStatusCompleted).
		Count(&count).Error
	return count, err
}

// CreateDelegation 创建委托
func (r *WorkflowTaskRepositoryImpl) CreateDelegation(ctx context.Context, delegation *entity.WorkflowDelegation) error {
	return r.db.WithContext(ctx).Create(delegation).Error
}

// UpdateDelegation 更新委托
func (r *WorkflowTaskRepositoryImpl) UpdateDelegation(ctx context.Context, delegation *entity.WorkflowDelegation) error {
	return r.db.WithContext(ctx).Save(delegation).Error
}

// FindDelegationByID 根据ID查找委托
func (r *WorkflowTaskRepositoryImpl) FindDelegationByID(ctx context.Context, id string) (*entity.WorkflowDelegation, error) {
	var delegation entity.WorkflowDelegation
	err := r.db.WithContext(ctx).First(&delegation, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &delegation, err
}

// FindActiveDelegation 查找有效的委托
func (r *WorkflowTaskRepositoryImpl) FindActiveDelegation(ctx context.Context, taskID string) (*entity.WorkflowDelegation, error) {
	var delegation entity.WorkflowDelegation
	err := r.db.WithContext(ctx).
		Where("task_id = ? AND status = 'active'", taskID).
		Where("(end_time IS NULL OR end_time > NOW())").
		First(&delegation).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &delegation, err
}

// CreateReminder 创建催办
func (r *WorkflowTaskRepositoryImpl) CreateReminder(ctx context.Context, reminder *entity.WorkflowReminder) error {
	return r.db.WithContext(ctx).Create(reminder).Error
}

// UpdateReminder 更新催办
func (r *WorkflowTaskRepositoryImpl) UpdateReminder(ctx context.Context, reminder *entity.WorkflowReminder) error {
	return r.db.WithContext(ctx).Save(reminder).Error
}

// FindReminderByTaskID 根据任务ID查找催办
func (r *WorkflowTaskRepositoryImpl) FindReminderByTaskID(ctx context.Context, taskID string) (*entity.WorkflowReminder, error) {
	var reminder entity.WorkflowReminder
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		First(&reminder).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &reminder, err
}
