package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// Create 创建任务（兼容 PostgreSQL JSON/UUID 字段）
func (r *TaskRepositoryImpl) Create(ctx context.Context, task *entity.Task) error {
	sanitizeTaskOptionalFields(task)

	db := r.db.WithContext(ctx)
	if strings.TrimSpace(task.ResultPhotos) == "" {
		db = db.Omit("ResultPhotos")
	}

	return db.Create(task).Error
}

// Update 更新任务（兼容 PostgreSQL JSON/UUID 字段）
func (r *TaskRepositoryImpl) Update(ctx context.Context, task *entity.Task) error {
	sanitizeTaskOptionalFields(task)

	db := r.db.WithContext(ctx)
	if strings.TrimSpace(task.ResultPhotos) == "" {
		db = db.Omit("ResultPhotos")
	}

	return db.Save(task).Error
}

func sanitizeTaskOptionalFields(task *entity.Task) {
	if task.AssigneeID != nil && strings.TrimSpace(*task.AssigneeID) == "" {
		task.AssigneeID = nil
	}
	if task.MissingPersonID != nil && strings.TrimSpace(*task.MissingPersonID) == "" {
		task.MissingPersonID = nil
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

	// 排序（白名单校验防止 SQL 注入）
	order := "created_at DESC"
	if query.SortField != "" {
		allowedSortFields := map[string]bool{
			"created_at": true, "updated_at": true, "deadline": true,
			"priority": true, "status": true, "title": true,
		}
		if allowedSortFields[query.SortField] {
			sortOrder := "DESC"
			if query.SortOrder == "asc" || query.SortOrder == "ASC" {
				sortOrder = "ASC"
			}
			order = query.SortField + " " + sortOrder
		}
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

// AddFollowUp 添加任务跟进
func (r *TaskRepositoryImpl) AddFollowUp(ctx context.Context, followUp *entity.TaskFollowUp) error {
	return r.db.WithContext(ctx).Create(followUp).Error
}

// GetFollowUps 获取任务跟进列表
func (r *TaskRepositoryImpl) GetFollowUps(ctx context.Context, taskID string, pagination repository.Pagination) (*repository.PageResult[entity.TaskFollowUp], error) {
	var list []entity.TaskFollowUp
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.TaskFollowUp{}).Where("task_id = ?", taskID)
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := db.Order("created_at DESC").
		Preload("User").
		Preload("Reviewer").
		Offset((pagination.Page - 1) * pagination.PageSize).
		Limit(pagination.PageSize).
		Find(&list).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(list, total, pagination.Page, pagination.PageSize), nil
}

// FindFollowUpByID 根据ID获取任务跟进
func (r *TaskRepositoryImpl) FindFollowUpByID(ctx context.Context, taskID, followUpID string) (*entity.TaskFollowUp, error) {
	var followUp entity.TaskFollowUp
	err := r.db.WithContext(ctx).
		Where("id = ? AND task_id = ?", followUpID, taskID).
		Preload("User").
		Preload("Reviewer").
		First(&followUp).Error
	if err != nil {
		return nil, err
	}
	return &followUp, nil
}

// UpdateFollowUp 更新任务跟进
func (r *TaskRepositoryImpl) UpdateFollowUp(ctx context.Context, followUp *entity.TaskFollowUp) error {
	return r.db.WithContext(ctx).Save(followUp).Error
}

// AddFollowUpComment 添加任务跟进评论
func (r *TaskRepositoryImpl) AddFollowUpComment(ctx context.Context, comment *entity.TaskFollowUpComment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&entity.TaskFollowUp{}).
			Where("id = ?", comment.FollowUpID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

// GetFollowUpComments 获取任务跟进评论
func (r *TaskRepositoryImpl) GetFollowUpComments(ctx context.Context, taskID, followUpID string, pagination repository.Pagination) (*repository.PageResult[entity.TaskFollowUpComment], error) {
	var list []entity.TaskFollowUpComment
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.TaskFollowUpComment{}).
		Where("task_id = ? AND follow_up_id = ?", taskID, followUpID)
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := db.Order("created_at DESC").
		Preload("User").
		Offset((pagination.Page - 1) * pagination.PageSize).
		Limit(pagination.PageSize).
		Find(&list).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(list, total, pagination.Page, pagination.PageSize), nil
}

// GetStats 获取统计
func (r *TaskRepositoryImpl) GetStats(ctx context.Context, userID string) (*entity.TaskStats, error) {
	stats := &entity.TaskStats{}

	baseDB := r.db.WithContext(ctx).Model(&entity.Task{})

	// 总数
	if err := baseDB.Count(&stats.Total).Error; err != nil {
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
		if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("status = ?", status).Count(&count).Error; err != nil {
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
		if err := r.db.WithContext(ctx).Model(&entity.Task{}).Where("assignee_id = ?", userID).Count(&myTasks).Error; err != nil {
			return nil, err
		}
		stats.MyTasks = myTasks

		var myPending int64
		if err := r.db.WithContext(ctx).Model(&entity.Task{}).
			Where("assignee_id = ? AND status IN (?, ?, ?)", userID, entity.TaskStatusPending, entity.TaskStatusAssigned, entity.TaskStatusProcessing).
			Count(&myPending).Error; err != nil {
			return nil, err
		}
		stats.MyPending = myPending

		var myProcessing int64
		if err := r.db.WithContext(ctx).Model(&entity.Task{}).
			Where("assignee_id = ? AND status = ?", userID, entity.TaskStatusProcessing).
			Count(&myProcessing).Error; err != nil {
			return nil, err
		}
		stats.MyProcessing = myProcessing

		var myCompleted int64
		if err := r.db.WithContext(ctx).Model(&entity.Task{}).
			Where("assignee_id = ? AND status = ?", userID, entity.TaskStatusCompleted).
			Count(&myCompleted).Error; err != nil {
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

// CountByDateRange 按日期范围统计
func (r *TaskRepositoryImpl) CountByDateRange(ctx context.Context, start, end string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&count).Error
	return count, err
}

// CountCompletedByDateRange 按日期范围统计已完成任务
func (r *TaskRepositoryImpl) CountCompletedByDateRange(ctx context.Context, start, end string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Task{}).
		Where("status = ? AND completed_at >= ? AND completed_at <= ?",
			entity.TaskStatusCompleted, start, end).
		Count(&count).Error
	return count, err
}

// FindOverdueTasks 查找所有活跃且已逾期的任务（用于自动标记）
func (r *TaskRepositoryImpl) FindOverdueTasks(ctx context.Context) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.WithContext(ctx).
		Where("deadline < ? AND status NOT IN (?, ?, ?) AND deleted_at IS NULL",
			time.Now(),
			entity.TaskStatusCompleted,
			entity.TaskStatusCancelled,
			entity.TaskStatusOverdue).
		Find(&tasks).Error
	return tasks, err
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
