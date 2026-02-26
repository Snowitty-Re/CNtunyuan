package service

import (
	"context"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/google/uuid"
)

// DialectService 方言服务
type DialectService struct {
	dialectRepo *repository.DialectRepository
}

// NewDialectService 创建方言服务
func NewDialectService(dialectRepo *repository.DialectRepository) *DialectService {
	return &DialectService{dialectRepo: dialectRepo}
}

// CreateDialectRequest 创建方言请求
type CreateDialectRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	AudioURL    string    `json:"audio_url" binding:"required"`
	Duration    int       `json:"duration" binding:"required,min=15,max=20"`
	FileSize    int       `json:"file_size"`
	Format      string    `json:"format"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	District    string    `json:"district"`
	Town        string    `json:"town"`
	Village     string    `json:"village"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	Address     string    `json:"address"`
	RecordTime  time.Time `json:"record_time"`
	Weather     string    `json:"weather"`
	Device      string    `json:"device"`
	TagIDs      []string  `json:"tag_ids"`
	OrgID       uuid.UUID `json:"org_id"`
}

// Create 创建方言记录
func (s *DialectService) Create(ctx context.Context, req *CreateDialectRequest, collectorID uuid.UUID) (*model.Dialect, error) {
	dialect := &model.Dialect{
		Title:       req.Title,
		Description: req.Description,
		AudioURL:    req.AudioURL,
		Duration:    req.Duration,
		FileSize:    req.FileSize,
		Format:      req.Format,
		Province:    req.Province,
		City:        req.City,
		District:    req.District,
		Town:        req.Town,
		Village:     req.Village,
		Longitude:   req.Longitude,
		Latitude:    req.Latitude,
		Address:     req.Address,
		RecordTime:  &req.RecordTime,
		Weather:     req.Weather,
		Device:      req.Device,
		CollectorID: collectorID,
		OrgID:       req.OrgID,
		Status:      "active",
	}

	if err := s.dialectRepo.Create(ctx, dialect); err != nil {
		return nil, err
	}

	return dialect, nil
}

// GetByID 根据ID获取
func (s *DialectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Dialect, error) {
	return s.dialectRepo.GetByID(ctx, id)
}

// Update 更新
func (s *DialectService) Update(ctx context.Context, id uuid.UUID, req *CreateDialectRequest) (*model.Dialect, error) {
	dialect, err := s.dialectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		dialect.Title = req.Title
	}
	if req.Description != "" {
		dialect.Description = req.Description
	}
	if req.AudioURL != "" {
		dialect.AudioURL = req.AudioURL
	}
	if req.Duration > 0 {
		dialect.Duration = req.Duration
	}
	if req.Province != "" {
		dialect.Province = req.Province
	}
	if req.City != "" {
		dialect.City = req.City
	}
	if req.District != "" {
		dialect.District = req.District
	}
	if req.Town != "" {
		dialect.Town = req.Town
	}
	if req.Village != "" {
		dialect.Village = req.Village
	}
	if req.Address != "" {
		dialect.Address = req.Address
	}

	if err := s.dialectRepo.Update(ctx, dialect); err != nil {
		return nil, err
	}

	return dialect, nil
}

// Delete 删除
func (s *DialectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.dialectRepo.Delete(ctx, id)
}

// List 列表查询
func (s *DialectService) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.Dialect, int64, error) {
	return s.dialectRepo.List(ctx, page, pageSize, filters)
}

// Play 播放方言
func (s *DialectService) Play(ctx context.Context, id, userID uuid.UUID, ip string, duration int) error {
	// 增加播放次数
	if err := s.dialectRepo.IncrementPlayCount(ctx, id); err != nil {
		return err
	}

	// 记录播放日志
	log := &model.DialectPlayLog{
		DialectID: id,
		UserID:    userID,
		IP:        ip,
		Duration:  duration,
	}

	return s.dialectRepo.AddPlayLog(ctx, log)
}

// Like 点赞
func (s *DialectService) Like(ctx context.Context, id, userID uuid.UUID) error {
	// 检查是否已点赞
	isLiked, err := s.dialectRepo.IsLiked(ctx, id, userID)
	if err != nil {
		return err
	}

	if isLiked {
		return nil // 已点赞，直接返回
	}

	// 添加点赞
	like := &model.DialectLike{
		DialectID: id,
		UserID:    userID,
	}
	if err := s.dialectRepo.AddLike(ctx, like); err != nil {
		return err
	}

	// 更新点赞数
	return s.dialectRepo.UpdateLikeCount(ctx, id)
}

// Unlike 取消点赞
func (s *DialectService) Unlike(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.dialectRepo.RemoveLike(ctx, id, userID); err != nil {
		return err
	}
	return s.dialectRepo.UpdateLikeCount(ctx, id)
}

// GetNearbyDialects 获取附近方言
func (s *DialectService) GetNearbyDialects(ctx context.Context, lat, lng, radius float64) ([]*model.Dialect, error) {
	return s.dialectRepo.GetNearbyDialects(ctx, lat, lng, radius)
}

// GetStatistics 获取统计
func (s *DialectService) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 总数量
	_, total, err := s.dialectRepo.List(ctx, 1, 1, nil)
	if err != nil {
		return nil, err
	}
	result["total"] = total

	// 今日新增
	today := time.Now().Format("2006-01-02")
	_, todayCount, err := s.dialectRepo.List(ctx, 1, 1, map[string]interface{}{
		"DATE(created_at)": today,
	})
	if err != nil {
		return nil, err
	}
	result["today"] = todayCount

	return result, nil
}
