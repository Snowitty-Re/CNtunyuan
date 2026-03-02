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
	ErrMissingPersonNotFound = errors.New("missing person not found")
	ErrInvalidStatus         = errors.New("invalid status")
)

// MissingPersonAppService 走失人员应用服务
type MissingPersonAppService struct {
	mpRepo repository.MissingPersonRepository
}

// NewMissingPersonAppService 创建走失人员应用服务
func NewMissingPersonAppService(mpRepo repository.MissingPersonRepository) *MissingPersonAppService {
	return &MissingPersonAppService{mpRepo: mpRepo}
}

// Create 创建走失人员
func (s *MissingPersonAppService) Create(ctx context.Context, req *dto.CreateMissingPersonRequest, reporterID string, orgID string) (*dto.MissingPersonResponse, error) {
	mp := &entity.MissingPerson{
		Name:         req.Name,
		Gender:       req.Gender,
		BirthDate:    &req.BirthDate,
		Age:          req.Age,
		Height:       req.Height,
		Weight:       req.Weight,
		Description:  req.Description,
		PhotoUrl:     req.PhotoUrl,
		MissingTime:  req.MissingTime,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		Clothes:      req.Clothes,
		Features:     req.Features,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactRel:   req.ContactRel,
		AltContact:   req.AltContact,
		ReporterID:   reporterID,
		OrgID:        orgID,
		Status:       entity.MissingStatusMissing,
		Urgency:      entity.UrgencyLevel(req.UrgencyLevel),
	}

	if req.UrgencyLevel == "" {
		mp.Urgency = entity.UrgencyLevelMedium
	}

	if err := s.mpRepo.Create(ctx, mp); err != nil {
		logger.Error("Failed to create missing person", logger.Err(err))
		return nil, err
	}

	logger.Info("Missing person created", logger.String("mp_id", mp.ID))

	resp := dto.ToMissingPersonResponse(mp)
	return &resp, nil
}

// GetByID 根据ID获取
func (s *MissingPersonAppService) GetByID(ctx context.Context, id string) (*dto.MissingPersonResponse, error) {
	mp, err := s.mpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrMissingPersonNotFound
	}

	// 增加浏览次数
	s.mpRepo.IncrementViews(ctx, id)

	resp := dto.ToMissingPersonResponse(mp)
	return &resp, nil
}

// List 列表查询
func (s *MissingPersonAppService) List(ctx context.Context, req *dto.MissingPersonListRequest) (*dto.MissingPersonListResponse, error) {
	query := repository.NewMissingPersonQuery()
	query.Page = req.Page
	query.PageSize = req.PageSize
	query.Keyword = req.Keyword
	query.Status = entity.MissingStatus(req.Status)
	query.Gender = req.Gender
	query.AgeMin = req.AgeMin
	query.AgeMax = req.AgeMax
	query.Province = req.Province
	query.City = req.City
	query.District = req.District
	query.UrgencyLevel = req.UrgencyLevel

	result, err := s.mpRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]dto.MissingPersonResponse, len(result.List))
	for i, mp := range result.List {
		list[i] = dto.ToMissingPersonResponse(&mp)
	}

	resp := dto.NewMissingPersonListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}

// Update 更新
func (s *MissingPersonAppService) Update(ctx context.Context, id string, req *dto.UpdateMissingPersonRequest) (*dto.MissingPersonResponse, error) {
	mp, err := s.mpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrMissingPersonNotFound
	}

	// 检查是否可以更新
	if !mp.CanUpdate() {
		return nil, errors.New("cannot update closed case")
	}

	if req.Name != "" {
		mp.Name = req.Name
	}
	if req.Gender != "" {
		mp.Gender = req.Gender
	}
	if !req.BirthDate.IsZero() {
		mp.BirthDate = &req.BirthDate
	}
	if req.Age > 0 {
		mp.Age = req.Age
	}
	if req.Height > 0 {
		mp.Height = req.Height
	}
	if req.Weight > 0 {
		mp.Weight = req.Weight
	}
	if req.Description != "" {
		mp.Description = req.Description
	}
	if req.PhotoUrl != "" {
		mp.PhotoUrl = req.PhotoUrl
	}
	if !req.MissingTime.IsZero() {
		mp.MissingTime = req.MissingTime
	}
	if req.Province != "" {
		mp.Province = req.Province
	}
	if req.City != "" {
		mp.City = req.City
	}
	if req.District != "" {
		mp.District = req.District
	}
	if req.Address != "" {
		mp.Address = req.Address
	}
	if req.Clothes != "" {
		mp.Clothes = req.Clothes
	}
	if req.Features != "" {
		mp.Features = req.Features
	}
	if req.ContactName != "" {
		mp.ContactName = req.ContactName
	}
	if req.ContactPhone != "" {
		mp.ContactPhone = req.ContactPhone
	}
	if req.ContactRel != "" {
		mp.ContactRel = req.ContactRel
	}
	if req.AltContact != "" {
		mp.AltContact = req.AltContact
	}
	if req.UrgencyLevel != "" {
		mp.Urgency = entity.UrgencyLevel(req.UrgencyLevel)
	}

	if err := s.mpRepo.Update(ctx, mp); err != nil {
		logger.Error("Failed to update missing person", logger.Err(err))
		return nil, err
	}

	resp := dto.ToMissingPersonResponse(mp)
	return &resp, nil
}

// Delete 删除
func (s *MissingPersonAppService) Delete(ctx context.Context, id string) error {
	if err := s.mpRepo.SoftDelete(ctx, id); err != nil {
		logger.Error("Failed to delete missing person", logger.Err(err))
		return err
	}
	return nil
}

// UpdateStatus 更新状态
func (s *MissingPersonAppService) UpdateStatus(ctx context.Context, id string, status string) error {
	mp, err := s.mpRepo.FindByID(ctx, id)
	if err != nil {
		return ErrMissingPersonNotFound
	}

	if !mp.CanUpdate() {
		return errors.New("cannot update closed case")
	}

	newStatus := entity.MissingStatus(status)
	mp.Status = newStatus

	if err := s.mpRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		return err
	}

	return nil
}

// MarkFound 标记找到
func (s *MissingPersonAppService) MarkFound(ctx context.Context, id string, req *dto.MarkFoundRequest) error {
	mp, err := s.mpRepo.FindByID(ctx, id)
	if err != nil {
		return ErrMissingPersonNotFound
	}

	if err := mp.MarkFound(req.Location, req.Note); err != nil {
		return err
	}

	// 更新数据库
	if err := s.mpRepo.Update(ctx, mp); err != nil {
		return err
	}

	return nil
}

// MarkReunited 标记团聚
func (s *MissingPersonAppService) MarkReunited(ctx context.Context, id string) error {
	mp, err := s.mpRepo.FindByID(ctx, id)
	if err != nil {
		return ErrMissingPersonNotFound
	}

	if err := mp.MarkReunited(); err != nil {
		return err
	}

	if err := s.mpRepo.Update(ctx, mp); err != nil {
		return err
	}

	return nil
}

// AddTrack 添加轨迹
func (s *MissingPersonAppService) AddTrack(ctx context.Context, personID string, req *dto.CreateMissingPersonTrackRequest, reporterID string) (*dto.MissingPersonTrackResponse, error) {
	// 检查案件是否存在
	if _, err := s.mpRepo.FindByID(ctx, personID); err != nil {
		return nil, ErrMissingPersonNotFound
	}

	track := &entity.MissingPersonTrack{
		MissingPersonID: personID,
		ReporterID:      reporterID,
		Location:        req.Location,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		Address:         req.Address,
		Time:            req.Time,
		Description:     req.Description,
		IsKeyPoint:      req.IsKeyPoint,
		Lat:             req.Lat,
		Lng:             req.Lng,
		Status:          "pending",
	}

	if err := s.mpRepo.AddTrack(ctx, track); err != nil {
		return nil, err
	}

	resp := dto.ToMissingPersonTrackResponse(track)
	return &resp, nil
}

// GetTracks 获取轨迹
func (s *MissingPersonAppService) GetTracks(ctx context.Context, personID string) ([]dto.MissingPersonTrackResponse, error) {
	tracks, err := s.mpRepo.GetTracks(ctx, personID)
	if err != nil {
		return nil, err
	}

	list := make([]dto.MissingPersonTrackResponse, len(tracks))
	for i, track := range tracks {
		list[i] = dto.ToMissingPersonTrackResponse(&track)
	}

	return list, nil
}

// GetStats 获取统计
func (s *MissingPersonAppService) GetStats(ctx context.Context) (*dto.MissingPersonStatsResponse, error) {
	stats, err := s.mpRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.MissingPersonStatsResponse{
		Total:     stats.Total,
		Missing:   stats.Missing,
		Searching: stats.Searching,
		Found:     stats.Found,
		Reunited:  stats.Reunited,
		Closed:    stats.Closed,
	}, nil
}

// Search 搜索
func (s *MissingPersonAppService) Search(ctx context.Context, keyword string, page, pageSize int) (*dto.MissingPersonListResponse, error) {
	pagination := repository.Pagination{Page: page, PageSize: pageSize}
	result, err := s.mpRepo.Search(ctx, keyword, pagination)
	if err != nil {
		return nil, err
	}

	list := make([]dto.MissingPersonResponse, len(result.List))
	for i, mp := range result.List {
		list[i] = dto.ToMissingPersonResponse(&mp)
	}

	resp := dto.NewMissingPersonListResponse(list, result.Total, result.Page, result.PageSize)
	return &resp, nil
}
