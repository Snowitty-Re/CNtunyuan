package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// WorkflowRepositoryImpl 工作流仓储实现
type WorkflowRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流仓储
func NewWorkflowRepository(db *gorm.DB) repository.WorkflowRepository {
	return &WorkflowRepositoryImpl{db: db}
}

// CreateDefinition 创建工作流定义
func (r *WorkflowRepositoryImpl) CreateDefinition(ctx context.Context, def *entity.WorkflowDefinition) error {
	return r.db.WithContext(ctx).Create(def).Error
}

// UpdateDefinition 更新工作流定义
func (r *WorkflowRepositoryImpl) UpdateDefinition(ctx context.Context, def *entity.WorkflowDefinition) error {
	return r.db.WithContext(ctx).Save(def).Error
}

// DeleteDefinition 删除工作流定义
func (r *WorkflowRepositoryImpl) DeleteDefinition(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkflowDefinition{}, "id = ?", id).Error
}

// FindDefinitionByID 根据ID查找工作流定义
func (r *WorkflowRepositoryImpl) FindDefinitionByID(ctx context.Context, id string) (*entity.WorkflowDefinition, error) {
	var def entity.WorkflowDefinition
	err := r.db.WithContext(ctx).
		Preload("Nodes").
		First(&def, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &def, err
}

// FindDefinitionByKey 根据Key和版本查找
func (r *WorkflowRepositoryImpl) FindDefinitionByKey(ctx context.Context, key string, version int) (*entity.WorkflowDefinition, error) {
	var def entity.WorkflowDefinition
	err := r.db.WithContext(ctx).
		Preload("Nodes").
		Where("key = ? AND version = ?", key, version).
		First(&def).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &def, err
}

// FindActiveDefinitionByKey 查找激活的版本
func (r *WorkflowRepositoryImpl) FindActiveDefinitionByKey(ctx context.Context, key string) (*entity.WorkflowDefinition, error) {
	var def entity.WorkflowDefinition
	err := r.db.WithContext(ctx).
		Preload("Nodes").
		Where("key = ? AND status = ?", key, entity.WorkflowStatusActive).
		Order("version DESC").
		First(&def).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &def, err
}

// ListDefinitions 列表查询
func (r *WorkflowRepositoryImpl) ListDefinitions(ctx context.Context, orgID, category string, page, pageSize int) ([]*entity.WorkflowDefinition, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.WorkflowDefinition{})
	
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	var defs []*entity.WorkflowDefinition
	err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&defs).Error
	
	return defs, total, err
}

// GetLatestVersion 获取最新版本号
func (r *WorkflowRepositoryImpl) GetLatestVersion(ctx context.Context, key string) (int, error) {
	var version int
	err := r.db.WithContext(ctx).
		Model(&entity.WorkflowDefinition{}).
		Where("key = ?", key).
		Select("COALESCE(MAX(version), 0)").
		Scan(&version).Error
	return version, err
}

// CreateNode 创建节点
func (r *WorkflowRepositoryImpl) CreateNode(ctx context.Context, node *entity.WorkflowNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

// UpdateNode 更新节点
func (r *WorkflowRepositoryImpl) UpdateNode(ctx context.Context, node *entity.WorkflowNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

// DeleteNode 删除节点
func (r *WorkflowRepositoryImpl) DeleteNode(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkflowNode{}, "id = ?", id).Error
}

// FindNodeByID 根据ID查找节点
func (r *WorkflowRepositoryImpl) FindNodeByID(ctx context.Context, id string) (*entity.WorkflowNode, error) {
	var node entity.WorkflowNode
	err := r.db.WithContext(ctx).First(&node, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &node, err
}

// FindNodesByWorkflowID 查找工作流的所有节点
func (r *WorkflowRepositoryImpl) FindNodesByWorkflowID(ctx context.Context, workflowID string) ([]*entity.WorkflowNode, error) {
	var nodes []*entity.WorkflowNode
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("order_index ASC").
		Find(&nodes).Error
	return nodes, err
}

// DeleteNodesByWorkflowID 删除工作流的所有节点
func (r *WorkflowRepositoryImpl) DeleteNodesByWorkflowID(ctx context.Context, workflowID string) error {
	return r.db.WithContext(ctx).
		Delete(&entity.WorkflowNode{}, "workflow_id = ?", workflowID).Error
}

// CreateInstance 创建流程实例
func (r *WorkflowRepositoryImpl) CreateInstance(ctx context.Context, instance *entity.WorkflowInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

// UpdateInstance 更新流程实例
func (r *WorkflowRepositoryImpl) UpdateInstance(ctx context.Context, instance *entity.WorkflowInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

// DeleteInstance 删除流程实例
func (r *WorkflowRepositoryImpl) DeleteInstance(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkflowInstance{}, "id = ?", id).Error
}

// FindInstanceByID 根据ID查找流程实例
func (r *WorkflowRepositoryImpl) FindInstanceByID(ctx context.Context, id string) (*entity.WorkflowInstance, error) {
	var instance entity.WorkflowInstance
	err := r.db.WithContext(ctx).
		Preload("Definition.Nodes").
		First(&instance, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &instance, err
}

// FindInstanceByBusinessKey 根据业务Key查找
func (r *WorkflowRepositoryImpl) FindInstanceByBusinessKey(ctx context.Context, businessKey string) (*entity.WorkflowInstance, error) {
	var instance entity.WorkflowInstance
	err := r.db.WithContext(ctx).
		Where("business_key = ?", businessKey).
		First(&instance).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &instance, err
}

// ListInstances 列表查询
func (r *WorkflowRepositoryImpl) ListInstances(ctx context.Context, query *entity.WorkflowQuery) ([]*entity.WorkflowInstance, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.WorkflowInstance{})
	
	if query.DefinitionID != "" {
		db = db.Where("definition_id = ?", query.DefinitionID)
	}
	if query.BusinessKey != "" {
		db = db.Where("business_key = ?", query.BusinessKey)
	}
	if query.BusinessID != "" {
		db = db.Where("business_id = ?", query.BusinessID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartedBy != "" {
		db = db.Where("started_by = ?", query.StartedBy)
	}
	if query.CurrentNodeID != "" {
		db = db.Where("current_node_id = ?", query.CurrentNodeID)
	}
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
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
	
	var instances []*entity.WorkflowInstance
	err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&instances).Error
	
	return instances, total, err
}

// ListMyInstances 查询我的流程实例
func (r *WorkflowRepositoryImpl) ListMyInstances(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowInstance, int64, error) {
	query := &entity.WorkflowQuery{
		StartedBy: userID,
		Page:      page,
		PageSize:  pageSize,
	}
	return r.ListInstances(ctx, query)
}

// ListMyTodoInstances 查询我的待办流程
func (r *WorkflowRepositoryImpl) ListMyTodoInstances(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowInstance, int64, error) {
	// 查询用户有待办任务的流程实例
	db := r.db.WithContext(ctx).
		Model(&entity.WorkflowInstance{}).
		Joins("JOIN ty_workflow_tasks ON ty_workflow_tasks.workflow_instance_id = ty_workflow_instances.id").
		Where("ty_workflow_tasks.assignee_id = ?", userID).
		Where("ty_workflow_tasks.status IN ?", []entity.TaskStatus{entity.TaskStatusPending, entity.TaskStatusProcessing})
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	var instances []*entity.WorkflowInstance
	err := db.Distinct("ty_workflow_instances.*").
		Order("ty_workflow_instances.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&instances).Error
	
	return instances, total, err
}

// ListMyDoneInstances 查询我的已办流程
func (r *WorkflowRepositoryImpl) ListMyDoneInstances(ctx context.Context, userID string, page, pageSize int) ([]*entity.WorkflowInstance, int64, error) {
	db := r.db.WithContext(ctx).
		Model(&entity.WorkflowInstance{}).
		Joins("JOIN ty_workflow_tasks ON ty_workflow_tasks.workflow_instance_id = ty_workflow_instances.id").
		Where("ty_workflow_tasks.assignee_id = ?", userID).
		Where("ty_workflow_tasks.status = ?", entity.TaskStatusCompleted)
	
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	var instances []*entity.WorkflowInstance
	err := db.Distinct("ty_workflow_instances.*").
		Order("ty_workflow_instances.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&instances).Error
	
	return instances, total, err
}

// CreateTransition 创建转换记录
func (r *WorkflowRepositoryImpl) CreateTransition(ctx context.Context, transition *entity.WorkflowTransition) error {
	return r.db.WithContext(ctx).Create(transition).Error
}

// FindTransitionsByInstanceID 查找实例的转换记录
func (r *WorkflowRepositoryImpl) FindTransitionsByInstanceID(ctx context.Context, instanceID string) ([]*entity.WorkflowTransition, error) {
	var transitions []*entity.WorkflowTransition
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Order("created_at ASC").
		Find(&transitions).Error
	return transitions, err
}

// GetInstanceStats 获取实例统计
func (r *WorkflowRepositoryImpl) GetInstanceStats(ctx context.Context, orgID string) (*entity.WorkflowStats, error) {
	db := r.db.WithContext(ctx).Model(&entity.WorkflowInstance{})
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	
	var stats entity.WorkflowStats
	
	// 总数
	db.Count(&stats.TotalInstances)
	
	// 各状态统计
	db.Where("status = ?", entity.InstanceStatusPending).Count(&stats.PendingCount)
	db.Where("status = ?", entity.InstanceStatusProcessing).Count(&stats.ProcessingCount)
	db.Where("status = ?", entity.InstanceStatusApproved).Count(&stats.ApprovedCount)
	db.Where("status = ?", entity.InstanceStatusRejected).Count(&stats.RejectedCount)
	db.Where("status = ?", entity.InstanceStatusCancelled).Count(&stats.CancelledCount)
	
	// 按流程定义统计
	var defStats []struct {
		DefinitionID string
		Count        int64
	}
	r.db.WithContext(ctx).
		Model(&entity.WorkflowInstance{}).
		Select("definition_id, COUNT(*) as count").
		Group("definition_id").
		Scan(&defStats)
	
	stats.DefinitionStats = make(map[string]int64)
	for _, s := range defStats {
		stats.DefinitionStats[s.DefinitionID] = s.Count
	}
	
	return &stats, nil
}

// GetUserTaskStats 获取用户任务统计
func (r *WorkflowRepositoryImpl) GetUserTaskStats(ctx context.Context, userID string) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	// 待办数量
	var todoCount int64
	r.db.WithContext(ctx).
		Model(&entity.WorkflowTask{}).
		Where("assignee_id = ? AND status IN ?", userID, []entity.TaskStatus{entity.TaskStatusPending, entity.TaskStatusProcessing}).
		Count(&todoCount)
	stats["todo"] = todoCount
	
	// 已办数量
	var doneCount int64
	r.db.WithContext(ctx).
		Model(&entity.WorkflowTask{}).
		Where("assignee_id = ? AND status = ?", userID, entity.TaskStatusCompleted).
		Count(&doneCount)
	stats["done"] = doneCount
	
	return stats, nil
}
