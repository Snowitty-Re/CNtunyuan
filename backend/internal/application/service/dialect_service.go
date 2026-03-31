package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/Snowitty-Re/CNtunyuan/pkg/validator"
	"github.com/google/uuid"
)

var (
	ErrDialectNotFound      = errors.New("dialect not found")
	ErrAlreadyLiked         = errors.New("already liked")
	ErrNotLiked             = errors.New("not liked")
	ErrDialectForbidden     = errors.New("no permission to modify this dialect")
	ErrDialectInvalidStatus = errors.New("invalid dialect status")
)

// DialectAppService 方言应用服务
type DialectAppService struct {
	dialectRepo repository.DialectRepository
}

// NewDialectAppService 创建方言应用服务
func NewDialectAppService(dialectRepo repository.DialectRepository) *DialectAppService {
	return &DialectAppService{dialectRepo: dialectRepo}
}

func canViewDialect(d *entity.Dialect, isManager bool) bool {
	return isManager || d.Status == entity.DialectStatusActive
}

func (s *DialectAppService) getViewableDialect(ctx context.Context, id string, isManager bool) (*entity.Dialect, error) {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDialectNotFound
	}
	if !canViewDialect(d, isManager) {
		return nil, ErrDialectNotFound
	}
	return d, nil
}

// Create 创建方言
func (s *DialectAppService) Create(ctx context.Context, req *dto.CreateDialectRequest, uploaderID string, orgID string) (*dto.DialectResponse, error) {
	d := &entity.Dialect{
		Title:       req.Title,
		Content:     req.Content,
		Region:      req.Region,
		Province:    req.Province,
		City:        req.City,
		DialectType: entity.DialectType(req.DialectType),
		AudioUrl:    req.AudioUrl,
		Duration:    req.Duration,
		FileSize:    req.FileSize,
		Format:      req.Format,
		Tags:        req.Tags,
		Description: req.Description,
		UploaderID:  uploaderID,
		OrgID:       orgID,
		Status:      entity.DialectStatusPending,
	}

	if req.MissingPersonID != "" {
		d.MissingPersonID = &req.MissingPersonID
	}

	if req.DialectType == "" {
		d.DialectType = entity.DialectTypePhrase
	}

	if err := s.dialectRepo.Create(ctx, d); err != nil {
		logger.Error("Failed to create dialect", logger.Err(err))
		return nil, err
	}

	logger.Info("Dialect created", logger.String("dialect_id", d.ID))

	resp := dto.ToDialectResponse(d, false)
	return &resp, nil
}

// GetByID 根据ID获取
func (s *DialectAppService) GetByID(ctx context.Context, id string, userID string, isManager bool) (*dto.DialectResponse, error) {
	d, err := s.getViewableDialect(ctx, id, isManager)
	if err != nil {
		return nil, err
	}

	isLiked := false
	if userID != "" {
		if liked, err := s.dialectRepo.HasLiked(ctx, id, userID); err == nil {
			isLiked = liked
		}
	}

	resp := dto.ToDialectResponse(d, isLiked)
	return &resp, nil
}

// List 列表查询
func (s *DialectAppService) List(ctx context.Context, req *dto.DialectListRequest, isManager bool) (*dto.DialectListResponse, error) {
	req.Page, req.PageSize = validator.SanitizePagination(req.Page, req.PageSize)

	query := repository.NewDialectQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Region = req.Region
	query.Province = req.Province
	query.City = req.City
	query.Type = entity.DialectType(req.Type)
	if isManager {
		query.Status = entity.DialectStatus(req.Status)
	} else {
		query.Status = entity.DialectStatusActive
	}
	query.SortBy = req.SortBy
	query.SortOrder = req.SortOrder

	result, err := s.dialectRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.DialectResponse, len(result.List))
	for i, d := range result.List {
		list[i] = dto.ToDialectResponse(&d, false)
	}

	resp := dto.NewDialectListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// Update 更新
func (s *DialectAppService) Update(ctx context.Context, id string, req *dto.UpdateDialectRequest, userID string, isManager bool) (*dto.DialectResponse, error) {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDialectNotFound
	}

	// 只有上传者或管理员可以修改
	if d.UploaderID != userID && !isManager {
		return nil, ErrDialectForbidden
	}

	if req.Title != "" {
		d.Title = req.Title
	}
	if req.Content != "" {
		d.Content = req.Content
	}
	if req.Region != "" {
		d.Region = req.Region
	}
	if req.Province != "" {
		d.Province = req.Province
	}
	if req.City != "" {
		d.City = req.City
	}
	if req.DialectType != "" {
		d.DialectType = entity.DialectType(req.DialectType)
	}
	if req.Tags != "" {
		d.Tags = req.Tags
	}
	if req.Description != "" {
		d.Description = req.Description
	}

	if err := s.dialectRepo.Update(ctx, d); err != nil {
		logger.Error("Failed to update dialect", logger.Err(err))
		return nil, err
	}

	resp := dto.ToDialectResponse(d, false)
	return &resp, nil
}

// Delete 删除
func (s *DialectAppService) Delete(ctx context.Context, id string, userID string, isManager bool) error {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}

	// 只有上传者或管理员可以删除
	if d.UploaderID != userID && !isManager {
		return ErrDialectForbidden
	}

	if err := s.dialectRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete dialect", logger.Err(err))
		return err
	}
	return nil
}

// UpdateStatus 更新状态
func (s *DialectAppService) UpdateStatus(ctx context.Context, id string, status string) error {
	newStatus := entity.DialectStatus(status)
	if !entity.IsValidDialectStatus(newStatus) {
		return fmt.Errorf("%w: %s，合法值为 active/inactive/pending", ErrDialectInvalidStatus, status)
	}

	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}

	d.Status = newStatus
	if err := s.dialectRepo.Update(ctx, d); err != nil {
		logger.Error("Failed to update dialect status", logger.Err(err), logger.String("dialect_id", id))
		return err
	}

	return nil
}

// Feature 设为精选
func (s *DialectAppService) Feature(ctx context.Context, id string) error {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}

	d.Feature()
	return s.dialectRepo.Update(ctx, d)
}

// Unfeature 取消精选
func (s *DialectAppService) Unfeature(ctx context.Context, id string) error {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}

	d.Unfeature()
	return s.dialectRepo.Update(ctx, d)
}

// IncrementPlayCount 增加播放次数
func (s *DialectAppService) IncrementPlayCount(ctx context.Context, id string, isManager bool) error {
	if _, err := s.getViewableDialect(ctx, id, isManager); err != nil {
		return err
	}
	return s.dialectRepo.IncrementPlayCount(ctx, id)
}

// Like 点赞
func (s *DialectAppService) Like(ctx context.Context, dialectID string, userID string, isManager bool) error {
	// 检查方言是否存在且可见
	if _, err := s.getViewableDialect(ctx, dialectID, isManager); err != nil {
		return err
	}

	// 检查是否已点赞
	hasLiked, err := s.dialectRepo.HasLiked(ctx, dialectID, userID)
	if err != nil {
		logger.Error("Failed to check like status", logger.Err(err))
		return err
	}
	if hasLiked {
		return ErrAlreadyLiked
	}

	like := &entity.DialectLike{
		ID:        uuid.New().String(),
		DialectID: dialectID,
		UserID:    userID,
	}
	if err := s.dialectRepo.AddLike(ctx, like); err != nil {
		logger.Error("Failed to add like", logger.Err(err))
		return err
	}

	return nil
}

// Unlike 取消点赞
func (s *DialectAppService) Unlike(ctx context.Context, dialectID string, userID string, isManager bool) error {
	if _, err := s.getViewableDialect(ctx, dialectID, isManager); err != nil {
		return err
	}

	// 检查是否已点赞
	hasLiked, err := s.dialectRepo.HasLiked(ctx, dialectID, userID)
	if err != nil {
		return err
	}
	if !hasLiked {
		return ErrNotLiked
	}

	if err := s.dialectRepo.RemoveLike(ctx, dialectID, userID); err != nil {
		return err
	}

	return nil
}

// HasLiked 是否已点赞
func (s *DialectAppService) HasLiked(ctx context.Context, dialectID string, userID string) (bool, error) {
	return s.dialectRepo.HasLiked(ctx, dialectID, userID)
}

// AddComment 添加评论
func (s *DialectAppService) AddComment(ctx context.Context, dialectID string, req *dto.CreateDialectCommentRequest, userID string, isManager bool) (*dto.DialectCommentResponse, error) {
	// 验证方言是否存在且可见
	if _, err := s.getViewableDialect(ctx, dialectID, isManager); err != nil {
		return nil, err
	}

	comment := &entity.DialectComment{
		DialectID: dialectID,
		UserID:    userID,
		Content:   req.Content,
	}
	if req.ParentID != "" {
		comment.ParentID = &req.ParentID
	}

	if err := s.dialectRepo.AddComment(ctx, comment); err != nil {
		return nil, err
	}

	resp := dto.ToDialectCommentResponse(comment)
	return &resp, nil
}

// GetComments 获取评论
func (s *DialectAppService) GetComments(ctx context.Context, dialectID string, page, pageSize int, isManager bool) (*dto.PageResult[dto.DialectCommentResponse], error) {
	if _, err := s.getViewableDialect(ctx, dialectID, isManager); err != nil {
		return nil, err
	}

	page, pageSize = validator.SanitizePagination(page, pageSize)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.dialectRepo.GetComments(ctx, dialectID, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.DialectCommentResponse, len(result.List))
	for i, c := range result.List {
		list[i] = dto.ToDialectCommentResponse(&c)
	}

	totalPages := int(result.Total) / pageSize
	if int(result.Total)%pageSize > 0 {
		totalPages++
	}

	return &dto.PageResult[dto.DialectCommentResponse]{
		List:       list,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetFeatured 获取精选方言
func (s *DialectAppService) GetFeatured(ctx context.Context, page, pageSize int) (*dto.DialectListResponse, error) {
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.dialectRepo.FindFeatured(ctx, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.DialectResponse, len(result.List))
	for i, d := range result.List {
		list[i] = dto.ToDialectResponse(&d, false)
	}

	resp := dto.NewDialectListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// GetStats 获取统计
func (s *DialectAppService) GetStats(ctx context.Context, isManager bool) (*dto.DialectStatsResponse, error) {
	stats, err := s.dialectRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	// 非管理角色仅返回公开数据（active）
	if !isManager {
		stats.Pending = 0
		stats.Total = stats.Active
	}

	return &dto.DialectStatsResponse{
		Total:      stats.Total,
		Active:     stats.Active,
		Pending:    stats.Pending,
		Featured:   stats.Featured,
		TotalPlays: stats.TotalPlays,
		TotalLikes: stats.TotalLikes,
	}, nil
}
