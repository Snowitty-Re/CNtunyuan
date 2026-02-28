package repository

import (
	"context"
	"fmt"
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

// GetByID 根据ID获取任务
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Assignee").
		Preload("MissingPerson").
		Preload("Org").
		Preload("Attachments").
		First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List 获取任务列表
func (r *TaskRepository) List(ctx context.Context, params map[string]interface{}) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Task{})

	// 应用过滤条件
	if status, ok := params["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if assigneeID, ok := params["assignee_id"].(string); ok && assigneeID != "" {
		query = query.Where("assignee_id = ?", assigneeID)
	}
	if creatorID, ok := params["creator_id"].(string); ok && creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}
	if orgID, ok := params["org_id"].(string); ok && orgID != "" {
		query = query.Where("org_id = ?", orgID)
	}
	if taskType, ok := params["type"].(string); ok && taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	if priority, ok := params["priority"].(string); ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if missingPersonID, ok := params["missing_person_id"].(string); ok && missingPersonID != "" {
		query = query.Where("missing_person_id = ?", missingPersonID)
	}
	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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
		Preload("Assignee").
		Preload("MissingPerson").
		Preload("Org").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks).Error

	return tasks, total, err
}

// Update 更新任务
func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// UpdateStatus 更新任务状态
func (r *TaskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 根据状态更新相关字段
	switch status {
	case model.TaskStatusAssigned:
		updates["start_time"] = time.Now()
	case model.TaskStatusCompleted:
		updates["completed_time"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete 删除任务
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, "id = ?", id).Error
}

// Assign 分配任务
func (r *TaskRepository) Assign(ctx context.Context, id uuid.UUID, assigneeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"assignee_id": assigneeID,
			"status":      model.TaskStatusAssigned,
			"start_time":  time.Now(),
		}).Error
}

// Unassign 取消分配
func (r *TaskRepository) Unassign(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"assignee_id": nil,
			"status":      model.TaskStatusPending,
		}).Error
}

// Transfer 转派任务
func (r *TaskRepository) Transfer(ctx context.Context, id uuid.UUID, fromUserID, toUserID uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ? AND assignee_id = ?", id, fromUserID).
		Updates(map[string]interface{}{
			"assignee_id": toUserID,
		}).Error
}

// Complete 完成任务
func (r *TaskRepository) Complete(ctx context.Context, id uuid.UUID, feedback, result string, attachments []string) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         model.TaskStatusCompleted,
			"feedback":       feedback,
			"result":         result,
			"completed_time": time.Now(),
			"progress":       100,
		}).Error
}

// Cancel 取消任务
func (r *TaskRepository) Cancel(ctx context.Context, id uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  model.TaskStatusCancelled,
			"result":  reason,
		}).Error
}

// UpdateProgress 更新进度
func (r *TaskRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Update("progress", progress).Error
}

// GetMyTasks 获取我的任务
func (r *TaskRepository) GetMyTasks(ctx context.Context, userID uuid.UUID, status string) ([]model.Task, error) {
	var tasks []model.Task
	query := r.db.WithContext(ctx).
		Where("assignee_id = ?", userID).
		Preload("Creator").
		Preload("MissingPerson").
		Order("created_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}

// GetCreatedTasks 获取我创建的任务
func (r *TaskRepository) GetCreatedTasks(ctx context.Context, userID uuid.UUID) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).
		Where("creator_id = ?", userID).
		Preload("Assignee").
		Preload("MissingPerson").
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// GetStatistics 获取任务统计
func (r *TaskRepository) GetStatistics(ctx context.Context, orgID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.WithContext(ctx).Model(&model.Task{})
	if orgID != "" {
		query = query.Where("org_id = ?", orgID)
	}

	// 总任务数
	var total int64
	query.Count(&total)
	stats["total"] = total

	// 各状态数量
	statuses := []string{
		model.TaskStatusDraft,
		model.TaskStatusPending,
		model.TaskStatusAssigned,
		model.TaskStatusProcessing,
		model.TaskStatusCompleted,
		model.TaskStatusCancelled,
	}
	for _, status := range statuses {
		var count int64
		query.Where("status = ?", status).Count(&count)
		stats[status] = count
	}

	// 今日新增
	var todayCount int64
	today := time.Now().Format("2006-01-02")
	query.Where("DATE(created_at) = ?", today).Count(&todayCount)
	stats["today"] = todayCount

	// 今日完成
	var todayCompleted int64
	query.Where("DATE(completed_time) = ? AND status = ?", today, model.TaskStatusCompleted).Count(&todayCompleted)
	stats["today_completed"] = todayCompleted

	// 逾期任务
	var overdueCount int64
	query.Where("deadline < ? AND status NOT IN ?", time.Now(), []string{model.TaskStatusCompleted, model.TaskStatusCancelled}).Count(&overdueCount)
	stats["overdue"] = overdueCount

	return stats, nil
}

// CreateLog 创建任务日志
func (r *TaskRepository) CreateLog(ctx context.Context, log *model.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLogs 获取任务日志
func (r *TaskRepository) GetLogs(ctx context.Context, taskID uuid.UUID) ([]model.TaskLog, error) {
	var logs []model.TaskLog
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Preload("User").
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// CreateComment 创建评论
func (r *TaskRepository) CreateComment(ctx context.Context, comment *model.TaskComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// GetComments 获取评论列表
func (r *TaskRepository) GetComments(ctx context.Context, taskID uuid.UUID) ([]model.TaskComment, error) {
	var comments []model.TaskComment
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Preload("User").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// AddAttachment 添加附件
func (r *TaskRepository) AddAttachment(ctx context.Context, attachment *model.TaskAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

// DeleteAttachment 删除附件
func (r *TaskRepository) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.TaskAttachment{}, "id = ?", id).Error
}

// BatchAssign 批量分配任务
func (r *TaskRepository) BatchAssign(ctx context.Context, taskIDs []uuid.UUID, assigneeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id IN ?", taskIDs).
		Updates(map[string]interface{}{
			"assignee_id": assigneeID,
			"status":      model.TaskStatusAssigned,
			"start_time":  time.Now(),
		}).Error
}

// GetPendingTasks 获取待分配任务（用于自动分配）
func (r *TaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).
		Where("status = ?", model.TaskStatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// AutoAssignByWorkload 根据工作量自动分配
func (r *TaskRepository) GetUserWorkload(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("assignee_id = ? AND status IN ?", userID, []string{model.TaskStatusAssigned, model.TaskStatusProcessing}).
		Count(&count).Error
	return count, err
}

// CheckTaskNoExists 检查任务编号是否存在
func (r *TaskRepository) CheckTaskNoExists(ctx context.Context, taskNo string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("task_no = ?", taskNo).
		Count(&count).Error
	return count > 0, err
}

// GenerateTaskNo 生成任务编号
func (r *TaskRepository) GenerateTaskNo(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		// 格式: TK + 年月日 + 6位随机字符
		taskNo := fmt.Sprintf("TK%s%s", time.Now().Format("20060102"), generateRandomString(6))
		exists, err := r.CheckTaskNoExists(ctx, taskNo)
		if err != nil {
			return "", err
		}
		if !exists {
			return taskNo, nil
		}
		// 如果存在，等待1毫秒后重试
		time.Sleep(time.Millisecond)
	}
	return "", fmt.Errorf("生成任务编号失败，请重试")
}

func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
