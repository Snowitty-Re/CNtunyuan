package service

import (
	"context"
	"errors"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

var (
	ErrDialectNotFound = errors.New("dialect not found")
	ErrAlreadyLiked    = errors.New("already liked")
	ErrNotLiked        = errors.New("not liked")
)

// DialectAppService 方言应用服务
type DialectAppService struct {
	dialectRepo repository.DialectRepository
}

// NewDialectAppService 创建方言应用服务
func NewDialectAppService(dialectRepo repository.DialectRepository) *DialectAppService {
	return &DialectAppService{dialectRepo: dialectRepo}
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

	if req.DialectType == "" {
		d.DialectType = entity.DialectTypePhrase
	}

	if err := s.dialectRepo.Create(ctx, d); err != nil {
		logger.Error("Failed to create dialect", logger.Err(err))
		return nil, err
	}

	logger.Info("Dialect created", logger.String("dialect_id", d.ID))

	resp := dto.ToDialectResponse(d)
	return &resp, nil
}

// GetByID 根据ID获取
func (s *DialectAppService) GetByID(ctx context.Context, id string) (*dto.DialectResponse, error) {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDialectNotFound
	}

	resp := dto.ToDialectResponse(d)
	return &resp, nil
}

// List 列表查询
func (s *DialectAppService) List(ctx context.Context, req *dto.DialectListRequest) (*dto.DialectListResponse, error) {
	query := repository.NewDialectQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Region = req.Region
	query.Province = req.Province
	query.City = req.City
	query.Type = entity.DialectType(req.Type)
	query.Status = entity.DialectStatus(req.Status)
	query.SortBy = req.SortBy
	query.SortOrder = req.SortOrder

	result, err := s.dialectRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.DialectResponse, len(result.List))
	for i, d := range result.List {
		list[i] = dto.ToDialectResponse(&d)
	}

	resp := dto.NewDialectListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// Update 更新
func (s *DialectAppService) Update(ctx context.Context, id string, req *dto.UpdateDialectRequest) (*dto.DialectResponse, error) {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDialectNotFound
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

	resp := dto.ToDialectResponse(d)
	return &resp, nil
}

// Delete 删除
func (s *DialectAppService) Delete(ctx context.Context, id string) error {
	if err := s.dialectRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete dialect", logger.Err(err))
		return err
	}
	return nil
}

// UpdateStatus 更新状态
func (s *DialectAppService) UpdateStatus(ctx context.Context, id string, status string) error {
	if _, err := s.dialectRepo.FindByID(ctx, id); err != nil {
		return ErrDialectNotFound
	}

	if err := s.dialectRepo.Update(ctx, &entity.Dialect{
		Status: entity.DialectStatus(status),
	}); err != nil {
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
func (s *DialectAppService) IncrementPlayCount(ctx context.Context, id string) error {
	return s.dialectRepo.IncrementPlayCount(ctx, id)
}

// Like 点赞
func (s *DialectAppService) Like(ctx context.Context, dialectID string, userID string) error {
	like := &entity.DialectLike{
		DialectID: dialectID,
		UserID:    userID,
	}
	return s.dialectRepo.AddLike(ctx, like)
}

// Unlike 取消点赞
func (s *DialectAppService) Unlike(ctx context.Context, dialectID string, userID string) error {
	return s.dialectRepo.RemoveLike(ctx, dialectID, userID)
}

// HasLiked 是否已点赞
func (s *DialectAppService) HasLiked(ctx context.Context, dialectID string, userID string) (bool, error) {
	return s.dialectRepo.HasLiked(ctx, dialectID, userID)
}

// AddComment 添加评论
func (s *DialectAppService) AddComment(ctx context.Context, dialectID string, req *dto.CreateDialectCommentRequest, userID string) (*dto.DialectCommentResponse, error) {
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
func (s *DialectAppService) GetComments(ctx context.Context, dialectID string, page, pageSize int) (*dto.PageResult[dto.DialectCommentResponse], error) {
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
		list[i] = dto.ToDialectResponse(&d)
	}

	resp := dto.NewDialectListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// GetStats 获取统计
func (s *DialectAppService) GetStats(ctx context.Context) (*dto.DialectStatsResponse, error) {
	stats, err := s.dialectRepo.GetStats(ctx)
	if err != nil {
		return nil, err
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
