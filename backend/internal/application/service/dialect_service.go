package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/application/dto"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	domainService "github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	apperrors "github.com/Snowitty-Re/CNtunyuan/pkg/errors"
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

const defaultOrgID = "00000000-0000-0000-0000-000000000000"

// DialectAppService 方言应用服务
type DialectAppService struct {
	dialectRepo repository.DialectRepository
	userRepo    repository.UserRepository
	orgRepo     repository.OrganizationRepository
	fileRepo    repository.FileRepository
	storage     domainService.StorageService
	authz       *AuthorizationService
}

// NewDialectAppService 创建方言应用服务
func NewDialectAppService(
	dialectRepo repository.DialectRepository,
	userRepo repository.UserRepository,
	orgRepo repository.OrganizationRepository,
	fileRepo repository.FileRepository,
	storage domainService.StorageService,
	authz ...*AuthorizationService,
) *DialectAppService {
	var authzSvc *AuthorizationService
	if len(authz) > 0 && authz[0] != nil {
		authzSvc = authz[0]
	} else {
		authzSvc = NewAuthorizationService(orgRepo)
	}
	return &DialectAppService{
		dialectRepo: dialectRepo,
		userRepo:    userRepo,
		orgRepo:     orgRepo,
		fileRepo:    fileRepo,
		storage:     storage,
		authz:       authzSvc,
	}
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
	normalizedOrgID := strings.TrimSpace(orgID)
	fallbackOrgID, fallbackErr := s.resolveUploaderOrgID(ctx, uploaderID)
	if normalizedOrgID == "" && fallbackOrgID != "" {
		normalizedOrgID = fallbackOrgID
	}
	if normalizedOrgID == "" {
		normalizedOrgID = defaultOrgID
	}

	dialectType := entity.DialectType(strings.TrimSpace(req.DialectType))
	if dialectType == "" {
		dialectType = entity.DialectTypePhrase
	}
	if !isValidDialectType(dialectType) {
		return nil, apperrors.New(apperrors.CodeInvalidParam, "方言类型不合法")
	}

	normalizedTags, err := normalizeDialectTags(req.Tags)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidParam, "标签格式不合法")
	}

	d := &entity.Dialect{
		Title:            strings.TrimSpace(req.Title),
		Content:          strings.TrimSpace(req.Content),
		Region:           strings.TrimSpace(req.Region),
		Province:         strings.TrimSpace(req.Province),
		City:             strings.TrimSpace(req.City),
		DialectType:      dialectType,
		AudioUrl:         strings.TrimSpace(req.AudioUrl),
		Duration:         req.Duration,
		FileSize:         req.FileSize,
		Format:           strings.ToLower(strings.TrimSpace(req.Format)),
		Tags:             normalizedTags,
		Description:      strings.TrimSpace(req.Description),
		CollectAddress:   strings.TrimSpace(req.CollectAddress),
		CollectLatitude:  req.CollectLatitude,
		CollectLongitude: req.CollectLongitude,
		UploaderID:       uploaderID,
		OrgID:            normalizedOrgID,
		Status:           entity.DialectStatusPending,
	}

	if req.MissingPersonID != "" {
		d.MissingPersonID = &req.MissingPersonID
	}

	if err := d.Validate(); err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidParam, err.Error())
	}
	if err := validateCollectLocation(d.CollectLatitude, d.CollectLongitude); err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidParam, err.Error())
	}

	if err := s.dialectRepo.Create(ctx, d); err != nil {
		if isDialectOrgForeignKeyError(err) && fallbackOrgID != "" && fallbackOrgID != d.OrgID {
			d.OrgID = fallbackOrgID
			if retryErr := s.dialectRepo.Create(ctx, d); retryErr == nil {
				logger.Info("Dialect created with uploader organization fallback", logger.String("dialect_id", d.ID), logger.String("org_id", d.OrgID))
				resp := dto.ToDialectResponse(d, false)
				return &resp, nil
			} else {
				err = retryErr
			}
		}

		logger.Error("Failed to create dialect", logger.Err(err))
		if mapped := mapDialectCreateError(err, fallbackErr); mapped != nil {
			return nil, mapped
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "create dialect failed")
	}

	logger.Info("Dialect created", logger.String("dialect_id", d.ID))

	resp := dto.ToDialectResponse(d, false)
	return &resp, nil
}

func validateCollectLocation(latitude, longitude float64) error {
	hasLat := latitude != 0
	hasLng := longitude != 0
	if !hasLat && !hasLng {
		return nil
	}
	if !hasLat || !hasLng {
		return errors.New("采集位置经纬度不完整")
	}
	if latitude < -90 || latitude > 90 {
		return errors.New("采集纬度超出范围")
	}
	if longitude < -180 || longitude > 180 {
		return errors.New("采集经度超出范围")
	}
	return nil
}

func (s *DialectAppService) resolveUploaderOrgID(ctx context.Context, uploaderID string) (string, error) {
	if s.userRepo == nil || strings.TrimSpace(uploaderID) == "" {
		return "", nil
	}
	user, err := s.userRepo.FindByID(ctx, uploaderID)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(user.OrgID), nil
}

func isDialectOrgForeignKeyError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "foreign key") &&
		(strings.Contains(errMsg, "fk_dialect_org") || strings.Contains(errMsg, "org_id"))
}

func mapDialectCreateError(err error, fallbackErr error) error {
	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "invalid input syntax for type uuid"):
		return apperrors.New(apperrors.CodeInvalidParam, "组织信息异常，请重新登录后重试")
	case strings.Contains(errMsg, "json"), strings.Contains(errMsg, "tags"):
		return apperrors.New(apperrors.CodeInvalidParam, "标签格式不合法")
	case strings.Contains(errMsg, "missing_person_id") && (strings.Contains(errMsg, "does not exist") || strings.Contains(errMsg, "unknown column")):
		return apperrors.New(apperrors.CodeInternal, "数据库结构未对齐，请执行最新迁移后重试")
	case (strings.Contains(errMsg, "collect_address") || strings.Contains(errMsg, "collect_latitude") || strings.Contains(errMsg, "collect_longitude")) &&
		(strings.Contains(errMsg, "does not exist") || strings.Contains(errMsg, "unknown column")):
		return apperrors.New(apperrors.CodeInternal, "数据库结构未对齐，请执行最新迁移后重试")
	case strings.Contains(errMsg, "foreign key") && (strings.Contains(errMsg, "fk_dialect_uploader") || strings.Contains(errMsg, "uploader_id")):
		return apperrors.New(apperrors.CodeInvalidParam, "上传用户不存在或已失效，请重新登录后重试")
	case strings.Contains(errMsg, "foreign key") && (strings.Contains(errMsg, "fk_dialect_org") || strings.Contains(errMsg, "org_id")):
		if fallbackErr != nil {
			return apperrors.New(apperrors.CodeInvalidParam, "组织信息异常，请联系管理员修复用户组织关系")
		}
		return apperrors.New(apperrors.CodeInvalidParam, "组织信息异常，请重新登录后重试")
	default:
		return nil
	}
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
func (s *DialectAppService) List(ctx context.Context, req *dto.DialectListRequest, operator *entity.User) (*dto.DialectListResponse, error) {
	req.Page, req.PageSize = validator.SanitizePagination(req.Page, req.PageSize)

	query := repository.NewDialectQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Region = req.Region
	query.Province = req.Province
	query.City = req.City
	query.Type = entity.DialectType(req.Type)
	isManager := operator != nil && entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager
	if isManager {
		query.Status = entity.DialectStatus(req.Status)
		if !operator.IsSuperAdmin() {
			orgIDs, err := collectManageableOrgIDs(ctx, s.orgRepo, operator)
			if err != nil {
				return nil, err
			}
			query.OrgIDs = orgIDs
		}
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
func (s *DialectAppService) Update(ctx context.Context, id string, req *dto.UpdateDialectRequest, operator *entity.User) (*dto.DialectResponse, error) {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDialectNotFound
	}
	userID := operator.ID
	isManager := entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager

	// 只有上传者或管理员可以修改
	if d.UploaderID != userID && !isManager {
		decision, scopeErr := s.authz.CanModifyDialect(ctx, operator, d)
		if scopeErr != nil {
			return nil, scopeErr
		}
		if !decision.Allowed {
			return nil, ErrDialectForbidden
		}
	}
	if d.UploaderID != userID && isManager && !operator.IsSuperAdmin() {
		decision, scopeErr := s.authz.CanModifyDialect(ctx, operator, d)
		if scopeErr != nil {
			return nil, scopeErr
		}
		if !decision.Allowed {
			return nil, ErrDialectForbidden
		}
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
		normalizedTags, err := normalizeDialectTags(req.Tags)
		if err != nil {
			return nil, apperrors.New(apperrors.CodeInvalidParam, "标签格式不合法")
		}
		d.Tags = normalizedTags
	}
	if req.Description != "" {
		d.Description = req.Description
	}

	if err := s.dialectRepo.Update(ctx, d); err != nil {
		logger.Error("Failed to update dialect", logger.Err(err))
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "json") || strings.Contains(errMsg, "tags") {
			return nil, apperrors.New(apperrors.CodeInvalidParam, "标签格式不合法")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "update dialect failed")
	}

	resp := dto.ToDialectResponse(d, false)
	return &resp, nil
}

func isValidDialectType(dialectType entity.DialectType) bool {
	switch dialectType {
	case entity.DialectTypePhrase, entity.DialectTypeStory, entity.DialectTypeSong, entity.DialectTypeDaily, entity.DialectTypeOther:
		return true
	default:
		return false
	}
}

func normalizeDialectTags(raw string) (string, error) {
	tags := strings.TrimSpace(raw)
	if tags == "" {
		return "", nil
	}

	// 已是合法 JSON（数组/对象/字符串）则直接使用
	var any interface{}
	if json.Unmarshal([]byte(tags), &any) == nil {
		return tags, nil
	}

	// 兼容普通字符串标签：按常见分隔符切分成 JSON 数组
	separators := []string{"，", ",", ";", "；", "|", " "}
	normalized := tags
	for _, sep := range separators {
		normalized = strings.ReplaceAll(normalized, sep, ",")
	}

	parts := strings.Split(normalized, ",")
	tagList := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		tagList = append(tagList, item)
	}
	if len(tagList) == 0 {
		return "", nil
	}

	buf, err := json.Marshal(tagList)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// Delete 删除
func (s *DialectAppService) Delete(ctx context.Context, id string, operator *entity.User) error {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}
	userID := operator.ID
	isManager := entity.GetRoleLevel(operator.Role) >= entity.RoleLevelManager

	// 只有上传者或管理员可以删除
	if d.UploaderID != userID && !isManager {
		decision, scopeErr := s.authz.CanModifyDialect(ctx, operator, d)
		if scopeErr != nil {
			return scopeErr
		}
		if !decision.Allowed {
			return ErrDialectForbidden
		}
	}
	if d.UploaderID != userID && isManager && !operator.IsSuperAdmin() {
		decision, scopeErr := s.authz.CanModifyDialect(ctx, operator, d)
		if scopeErr != nil {
			return scopeErr
		}
		if !decision.Allowed {
			return ErrDialectForbidden
		}
	}

	if err := s.dialectRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete dialect", logger.Err(err))
		return err
	}

	s.cleanupDialectAudio(ctx, d.AudioUrl)
	return nil
}

func (s *DialectAppService) cleanupDialectAudio(ctx context.Context, audioURL string) {
	audioURL = strings.TrimSpace(audioURL)
	if audioURL == "" {
		return
	}

	filePath := deriveStoragePathFromAudioURL(audioURL)

	if s.fileRepo != nil {
		file, err := s.fileRepo.FindByURLOrPath(ctx, audioURL, filePath)
		if err == nil && file != nil {
			if err := s.fileRepo.SoftDelete(ctx, file.ID); err != nil {
				logger.Warn("Failed to soft delete dialect audio file record", logger.String("file_id", file.ID), logger.Err(err))
			}
			if s.storage != nil {
				if err := s.storage.Delete(ctx, file.Path); err != nil {
					logger.Warn("Failed to delete dialect audio physical file", logger.String("path", file.Path), logger.Err(err))
				}
			}
			return
		}
	}

	if s.storage != nil && filePath != "" {
		if err := s.storage.Delete(ctx, filePath); err != nil {
			logger.Warn("Failed to delete dialect audio file by derived path", logger.String("path", filePath), logger.Err(err))
		}
	}
}

func deriveStoragePathFromAudioURL(audioURL string) string {
	if audioURL == "" {
		return ""
	}
	trimmed := strings.TrimSpace(audioURL)
	if parsed, err := url.Parse(trimmed); err == nil && parsed.Path != "" {
		trimmed = parsed.Path
	}
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimPrefix(trimmed, "uploads/")
	return trimmed
}

// UpdateStatus 更新状态
func (s *DialectAppService) UpdateStatus(ctx context.Context, id string, status string, operator *entity.User) error {
	newStatus := entity.DialectStatus(status)
	if !entity.IsValidDialectStatus(newStatus) {
		return fmt.Errorf("%w: %s，合法值为 active/inactive/pending", ErrDialectInvalidStatus, status)
	}

	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}
	if operator != nil && !operator.IsSuperAdmin() {
		decision, scopeErr := s.authz.CanManageDialectStatus(ctx, operator, d)
		if scopeErr != nil {
			return scopeErr
		}
		if !decision.Allowed {
			return ErrDialectForbidden
		}
	}

	d.Status = newStatus
	if err := s.dialectRepo.Update(ctx, d); err != nil {
		logger.Error("Failed to update dialect status", logger.Err(err), logger.String("dialect_id", id))
		return err
	}

	return nil
}

// Feature 设为精选
func (s *DialectAppService) Feature(ctx context.Context, id string, operator *entity.User) error {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}
	if operator != nil && !operator.IsSuperAdmin() {
		decision, scopeErr := s.authz.CanManageDialectStatus(ctx, operator, d)
		if scopeErr != nil {
			return scopeErr
		}
		if !decision.Allowed {
			return ErrDialectForbidden
		}
	}

	d.Feature()
	return s.dialectRepo.Update(ctx, d)
}

// Unfeature 取消精选
func (s *DialectAppService) Unfeature(ctx context.Context, id string, operator *entity.User) error {
	d, err := s.dialectRepo.FindByID(ctx, id)
	if err != nil {
		return ErrDialectNotFound
	}
	if operator != nil && !operator.IsSuperAdmin() {
		decision, scopeErr := s.authz.CanManageDialectStatus(ctx, operator, d)
		if scopeErr != nil {
			return scopeErr
		}
		if !decision.Allowed {
			return ErrDialectForbidden
		}
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
