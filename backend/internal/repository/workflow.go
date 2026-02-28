package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowRepository 工作流仓库
type WorkflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流仓库
func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Create 创建工作流
func (r *WorkflowRepository) Create(ctx context.Context, workflow *model.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

// GetByID 根据ID获取工作流
func (r *WorkflowRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Preload("Creator").
		First(&workflow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// GetByCode 根据编码获取工作流
func (r *WorkflowRepository) GetByCode(ctx context.Context, code string) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Where("code = ? AND status = ?", code, model.WorkflowStatusActive).
		First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// List 获取工作流列表
func (r *WorkflowRepository) List(ctx context.Context, params map[string]interface{}) ([]model.Workflow, int64, error) {
	var workflows []model.Workflow
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Workflow{})

	// 应用过滤条件
	if status, ok := params["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if workflowType, ok := params["type"].(string); ok && workflowType != "" {
		query = query.Where("type = ?", workflowType)
	}
	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := 1
	pageSize := 20
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	if ps, ok := params["page_size"].(int); ok && ps > 0 {
		pageSize = ps
	}
	offset := (page - 1) * pageSize

	// 查询数据
	err := query.Preload("Creator").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&workflows).Error

	return workflows, total, err
}

// Update 更新工作流
func (r *WorkflowRepository) Update(ctx context.Context, workflow *model.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

// Delete 删除工作流
func (r *WorkflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除关联的步骤
		if err := tx.Where("workflow_id = ?", id).Delete(&model.WorkflowStep{}).Error; err != nil {
			return err
		}
		// 删除工作流
		return tx.Delete(&model.Workflow{}, "id = ?", id).Error
	})
}

// CreateStep 创建工作流步骤
func (r *WorkflowRepository) CreateStep(ctx context.Context, step *model.WorkflowStep) error {
	return r.db.WithContext(ctx).Create(step).Error
}

// GetStepByID 根据ID获取步骤
func (r *WorkflowRepository) GetStepByID(ctx context.Context, id uuid.UUID) (*model.WorkflowStep, error) {
	var step model.WorkflowStep
	err := r.db.WithContext(ctx).First(&step, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

// UpdateStep 更新步骤
func (r *WorkflowRepository) UpdateStep(ctx context.Context, step *model.WorkflowStep) error {
	return r.db.WithContext(ctx).Save(step).Error
}

// DeleteStep 删除步骤
func (r *WorkflowRepository) DeleteStep(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.WorkflowStep{}, "id = ?", id).Error
}

// GetStepsByWorkflowID 获取工作流的所有步骤
func (r *WorkflowRepository) GetStepsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]model.WorkflowStep, error) {
	var steps []model.WorkflowStep
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("step_order ASC").
		Find(&steps).Error
	return steps, err
}

// ReorderSteps 重新排序步骤
func (r *WorkflowRepository) ReorderSteps(ctx context.Context, workflowID uuid.UUID, stepIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, stepID := range stepIDs {
			if err := tx.Model(&model.WorkflowStep{}).
				Where("id = ? AND workflow_id = ?", stepID, workflowID).
				Update("step_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateInstance 创建工作流实例
func (r *WorkflowRepository) CreateInstance(ctx context.Context, instance *model.WorkflowInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

// GetInstanceByID 根据ID获取实例
func (r *WorkflowRepository) GetInstanceByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	var instance model.WorkflowInstance
	err := r.db.WithContext(ctx).
		Preload("Workflow").
		Preload("CurrentStep").
		Preload("Starter").
		Preload("History", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("History.Operator").
		First(&instance, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetInstancesByBusinessID 根据业务ID获取实例
func (r *WorkflowRepository) GetInstancesByBusinessID(ctx context.Context, businessID uuid.UUID) ([]model.WorkflowInstance, error) {
	var instances []model.WorkflowInstance
	err := r.db.WithContext(ctx).
		Where("business_id = ?", businessID).
		Preload("Workflow").
		Preload("CurrentStep").
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

// UpdateInstance 更新实例
func (r *WorkflowRepository) UpdateInstance(ctx context.Context, instance *model.WorkflowInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

// CreateHistory 创建历史记录
func (r *WorkflowRepository) CreateHistory(ctx context.Context, history *model.WorkflowHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetInstanceHistory 获取实例历史
func (r *WorkflowRepository) GetInstanceHistory(ctx context.Context, instanceID uuid.UUID) ([]model.WorkflowHistory, error) {
	var histories []model.WorkflowHistory
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Preload("Operator").
		Order("created_at ASC").
		Find(&histories).Error
	return histories, err
}

// ListInstances 获取实例列表
func (r *WorkflowRepository) ListInstances(ctx context.Context, params map[string]interface{}) ([]model.WorkflowInstance, int64, error) {
	var instances []model.WorkflowInstance
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WorkflowInstance{})

	// 应用过滤条件
	if status, ok := params["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if workflowID, ok := params["workflow_id"].(string); ok && workflowID != "" {
		query = query.Where("workflow_id = ?", workflowID)
	}
	if starterID, ok := params["starter_id"].(string); ok && starterID != "" {
		query = query.Where("starter_id = ?", starterID)
	}
	if businessType, ok := params["business_type"].(string); ok && businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := 1
	pageSize := 20
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	if ps, ok := params["page_size"].(int); ok && ps > 0 {
		pageSize = ps
	}
	offset := (page - 1) * pageSize

	// 查询数据
	err := query.Preload("Workflow").
		Preload("CurrentStep").
		Preload("Starter").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&instances).Error

	return instances, total, err
}

// GetActiveInstances 获取进行中的实例
func (r *WorkflowRepository) GetActiveInstances(ctx context.Context) ([]model.WorkflowInstance, error) {
	var instances []model.WorkflowInstance
	err := r.db.WithContext(ctx).
		Where("status = ?", "running").
		Preload("Workflow").
		Preload("CurrentStep").
		Find(&instances).Error
	return instances, err
}

// GetDefaultWorkflow 获取默认工作流
func (r *WorkflowRepository) GetDefaultWorkflow(ctx context.Context, workflowType string) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Where("type = ? AND is_default = ? AND status = ?", workflowType, true, model.WorkflowStatusActive).
		First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}
