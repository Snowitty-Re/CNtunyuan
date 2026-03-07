package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// TaskRepositoryImpl 任务仓储实现
type TaskRepositoryImpl struct {
	*BaseRepository[entity.Task]
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(db *gorm.DB) repository.TaskRepository {
	return &TaskRepositoryImpl{
		BaseRepository: NewBaseRepository[entity.Task](db),
	}
}

// List 分页查询
func (r *TaskRepositoryImpl) List(ctx context.Context, query *repository.TaskQuery) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Task{})

	// 关键词搜索
	if query.Keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 类型筛选
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}

	// 状态筛选
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 优先级筛选
	if query.Priority != "" {
		db = db.Where("priority = ?", query.Priority)
	}

	// 创建人筛选
	if query.CreatorID != "" {
		db = db.Where("creator_id = ?", query.CreatorID)
	}

	// 执行人筛选
	if query.AssigneeID != "" {
		db = db.Where("assignee_id = ?", query.AssigneeID)
	}

	// 组织筛选
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
	}

	// 走失人员关联
	if query.MissingPersonID != "" {
		db = db.Where("missing_person_id = ?", query.MissingPersonID)
	}

	// 地区筛选
	if query.Province != "" {
		db = db.Where("province = ?", query.Province)
	}
	if query.City != "" {
		db = db.Where("city = ?", query.City)
	}

	// 日期范围
	if query.StartDate != "" {
		db = db.Where("created_at >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("created_at <= ?", query.EndDate)
	}

	// 逾期筛选
	if query.IsOverdue != nil && *query.IsOverdue {
		db = db.Where("deadline < ? AND status NOT IN (?, ?)", time.Now(), entity.TaskStatusCompleted, entity.TaskStatusCancelled)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	order := "created_at DESC"
	if query.SortField != "" {
		order = query.SortField + " " + query.SortOrder
	}

	// 分页查询
	if err := db.Order(order).
		Preload("Creator").
		Preload("Assignee").
		Preload("MissingPerson").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, query.Page, query.PageSize), nil
}

// FindByAssignee 根据执行人查找
func (r *TaskRepositoryImpl) FindByAssignee(ctx context.Context, assigneeID string, pagination repository.Pagination) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Where("assignee_id = ?", assigneeID)

	if err := db.Model(&entity.Task{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, pagination.Page, pagination.PageSize), nil
}

// FindByCreator 根据创建人查找
func (r *TaskRepositoryImpl) FindByCreator(ctx context.Context, creatorID string, pagination repository.Pagination) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Where("creator_id = ?", creatorID)

	if err := db.Model(&entity.Task{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, pagination.Page, pagination.PageSize), nil
}

// FindByStatus 根据状态查找
func (r *TaskRepositoryImpl) FindByStatus(ctx context.Context, status entity.TaskStatus, pagination repository.Pagination) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Where("status = ?", status)

	if err := db.Model(&entity.Task{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, pagination.Page, pagination.PageSize), nil
}

// FindByOrg 根据组织查找
func (r *TaskRepositoryImpl) FindByOrg(ctx context.Context, orgID string, pagination repository.Pagination) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Where("org_id = ?", orgID)

	if err := db.Model(&entity.Task{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, pagination.Page, pagination.PageSize), nil
}

// FindByMissingPerson 根据走失人员查找
func (r *TaskRepositoryImpl) FindByMissingPerson(ctx context.Context, missingPersonID string) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.WithContext(ctx).Where("missing_person_id = ?", missingPersonID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// FindPending 查找待分配任务
func (r *TaskRepositoryImpl) FindPending(ctx context.Context, pagination repository.Pagination) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Where("status = ?", entity.TaskStatusPending)

	if err := db.Model(&entity.Task{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, pagination.Page, pagination.PageSize), nil
}

// FindOverdue 查找逾期任务
func (r *TaskRepositoryImpl) FindOverdue(ctx context.Context, pagination repository.Pagination) (*repository.PageResult[entity.Task], error) {
	var tasks []entity.Task
	var total int64

	db := r.db.WithContext(ctx).Where("deadline < ? AND status NOT IN (?, ?)", time.Now(), entity.TaskStatusCompleted, entity.TaskStatusCancelled)

	if err := db.Model(&entity.Task{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Order("deadline ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(tasks, total, pagination.Page, pagination.PageSize), nil
}

// UpdateStatus 更新状态
func (r *TaskRepositoryImpl) UpdateStatus(ctx context.Context, id string, status entity.TaskStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 如果开始处理，记录开始时间
	if status == entity.TaskStatusProcessing {
		now := time.Now()
		updates["started_at"] = &now
	}

	// 如果完成，记录完成时间
	if status == entity.TaskStatusCompleted {
		now := time.Now()
		updates["completed_at"] = &now
		updates["progress"] = 100
	}

	return r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateProgress 更新进度
func (r *TaskRepositoryImpl) UpdateProgress(ctx context.Context, id string, progress int) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", id).Update("progress", progress).Error
}

// AddLog 添加日志
func (r *TaskRepositoryImpl) AddLog(ctx context.Context, log *entity.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLogs 获取日志
func (r *TaskRepositoryImpl) GetLogs(ctx context.Context, taskID string) ([]entity.TaskLog, error) {
	var logs []entity.TaskLog
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at DESC").Preload("User").Find(&logs).Error
	return logs, err
}

// AddAttachment 添加附件
func (r *TaskRepositoryImpl) AddAttachment(ctx context.Context, attachment *entity.TaskAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

// GetAttachments 获取附件
func (r *TaskRepositoryImpl) GetAttachments(ctx context.Context, taskID string) ([]entity.TaskAttachment, error) {
	var attachments []entity.TaskAttachment
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("created_at DESC").Find(&attachments).Error
	return attachments, err
}

// GetStats 获取统计
func (r *TaskRepositoryImpl) GetStats(ctx context.Context, userID string) (*entity.TaskStats, error) {
	stats := &entity.TaskStats{}

	db := r.db.WithContext(ctx).Model(&entity.Task{})

	// 总数
	if err := db.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 各状态统计
	statuses := []entity.TaskStatus{
		entity.TaskStatusDraft,
		entity.TaskStatusPending,
		entity.TaskStatusAssigned,
		entity.TaskStatusProcessing,
		entity.TaskStatusCompleted,
		entity.TaskStatusCancelled,
		entity.TaskStatusOverdue,
	}

	for _, status := range statuses {
		var count int64
		if err := db.Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}
		switch status {
		case entity.TaskStatusDraft:
			stats.Draft = count
		case entity.TaskStatusPending:
			stats.Pending = count
		case entity.TaskStatusAssigned:
			stats.Assigned = count
		case entity.TaskStatusProcessing:
			stats.Processing = count
		case entity.TaskStatusCompleted:
			stats.Completed = count
		case entity.TaskStatusCancelled:
			stats.Cancelled = count
		case entity.TaskStatusOverdue:
			stats.Overdue = count
		}
	}

	// 我的任务统计
	if userID != "" {
		var myTasks int64
		if err := db.Where("creator_id = ? OR assignee_id = ?", userID, userID).Count(&myTasks).Error; err != nil {
			return nil, err
		}
		stats.MyTasks = myTasks

		var myPending int64
		if err := db.Where("(creator_id = ? OR assignee_id = ?) AND status IN (?, ?, ?)",
			userID, userID, entity.TaskStatusPending, entity.TaskStatusAssigned, entity.TaskStatusProcessing).Count(&myPending).Error; err != nil {
			return nil, err
		}
		stats.MyPending = myPending

		var myCompleted int64
		if err := db.Where("(creator_id = ? OR assignee_id = ?) AND status = ?",
			userID, userID, entity.TaskStatusCompleted).Count(&myCompleted).Error; err != nil {
			return nil, err
		}
		stats.MyCompleted = myCompleted
	}

	return stats, nil
}

// CountByStatus 按状态统计
func (r *TaskRepositoryImpl) CountByStatus(ctx context.Context, status entity.TaskStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountByAssignee 按执行人统计
func (r *TaskRepositoryImpl) CountByAssignee(ctx context.Context, assigneeID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("assignee_id = ?", assigneeID).Count(&count).Error
	return count, err
}

// CountOverdue 统计逾期任务
func (r *TaskRepositoryImpl) CountOverdue(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("deadline < ? AND status NOT IN (?, ?)", time.Now(), entity.TaskStatusCompleted, entity.TaskStatusCancelled).
		Count(&count).Error
	return count, err
}

// FindByID 根据ID查找
func (r *TaskRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Task, error) {
	var task entity.Task
	err := r.db.WithContext(ctx).Preload("Creator").Preload("Assignee").Preload("MissingPerson").First(&task, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	return &task, nil
}
