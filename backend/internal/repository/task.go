package repository

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepository 任务仓库
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓库
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create 创建任务
func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).Preload("Creator").Preload("Assignee").Preload("Org").
		Preload("MissingPerson").Preload("Attachments").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByTaskNo 根据任务编号获取
func (r *TaskRepository) GetByTaskNo(ctx context.Context, taskNo string) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).Where("task_no = ?", taskNo).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新
func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, id).Error
}

// List 列表查询
func (r *TaskRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Task{})

	// 过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "title":
				query = query.Where("title LIKE ?", "%"+value.(string)+"%")
			case "status":
				query = query.Where("status = ?", value)
			case "type":
				query = query.Where("type = ?", value)
			case "priority":
				query = query.Where("priority = ?", value)
			case "creator_id":
				query = query.Where("creator_id = ?", value)
			case "assignee_id":
				query = query.Where("assignee_id = ?", value)
			case "org_id":
				query = query.Where("org_id = ?", value)
			case "missing_person_id":
				query = query.Where("missing_person_id = ?", value)
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Creator").Preload("Assignee").Preload("MissingPerson").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// UpdateStatus 更新状态
func (r *TaskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", id).Update("status", status).Error
}

// Assign 分配任务
func (r *TaskRepository) Assign(ctx context.Context, id, assigneeID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"assignee_id": assigneeID,
		"status":      model.TaskStatusAssigned,
	}).Error
}

// GetUserTasks 获取用户的任务
func (r *TaskRepository) GetUserTasks(ctx context.Context, userID uuid.UUID, status string, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Task{}).Where("assignee_id = ? OR creator_id = ?", userID, userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("MissingPerson").Preload("Creator").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetOverdueTasks 获取逾期任务
func (r *TaskRepository) GetOverdueTasks(ctx context.Context) ([]*model.Task, error) {
	var tasks []*model.Task
	now := time.Now()
	err := r.db.WithContext(ctx).Where("deadline < ? AND status NOT IN ?", 
		now, []string{model.TaskStatusCompleted, model.TaskStatusCancelled}).
		Find(&tasks).Error
	return tasks, err
}

// AddLog 添加日志
func (r *TaskRepository) AddLog(ctx context.Context, log *model.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLogs 获取日志
func (r *TaskRepository) GetLogs(ctx context.Context, taskID uuid.UUID) ([]*model.TaskLog, error) {
	var logs []*model.TaskLog
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Preload("User").Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// AddComment 添加评论
func (r *TaskRepository) AddComment(ctx context.Context, comment *model.TaskComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// GetComments 获取评论
func (r *TaskRepository) GetComments(ctx context.Context, taskID uuid.UUID, page, pageSize int) ([]*model.TaskComment, int64, error) {
	var comments []*model.TaskComment
	var total int64

	query := r.db.WithContext(ctx).Model(&model.TaskComment{}).Where("task_id = ?", taskID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, err
}

// AddAttachment 添加附件
func (r *TaskRepository) AddAttachment(ctx context.Context, attachment *model.TaskAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

// GetStatistics 获取任务统计
func (r *TaskRepository) GetStatistics(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var result = make(map[string]interface{})

	// 总任务数
	var total int64
	r.db.WithContext(ctx).Model(&model.Task{}).Where("org_id = ?", orgID).Count(&total)
	result["total"] = total

	// 各状态统计
	statuses := []string{model.TaskStatusPending, model.TaskStatusAssigned, model.TaskStatusProcessing, model.TaskStatusCompleted}
	for _, status := range statuses {
		var count int64
		r.db.WithContext(ctx).Model(&model.Task{}).Where("org_id = ? AND status = ?", orgID, status).Count(&count)
		result[status] = count
	}

	// 今日新增
	var todayCount int64
	today := time.Now().Format("2006-01-02")
	r.db.WithContext(ctx).Model(&model.Task{}).Where("org_id = ? AND DATE(created_at) = ?", orgID, today).Count(&todayCount)
	result["today"] = todayCount

	return result, nil
}
