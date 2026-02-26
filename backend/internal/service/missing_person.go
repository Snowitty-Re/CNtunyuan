package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/Snowitty-Re/CNtunyuan/internal/repository"
	"github.com/google/uuid"
)

// MissingPersonService 走失人员服务
type MissingPersonService struct {
	mpRepo     *repository.MissingPersonRepository
	orgRepo    *repository.OrganizationRepository
}

// NewMissingPersonService 创建服务
func NewMissingPersonService(mpRepo *repository.MissingPersonRepository, orgRepo *repository.OrganizationRepository) *MissingPersonService {
	return &MissingPersonService{
		mpRepo:  mpRepo,
		orgRepo: orgRepo,
	}
}

// CreateMissingPersonRequest 创建请求
type CreateMissingPersonRequest struct {
	Name            string     `json:"name" binding:"required"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Age             int        `json:"age"`
	Height          int        `json:"height"`
	Weight          int        `json:"weight"`
	IDCard          string     `json:"id_card"`
	Appearance      string     `json:"appearance"`
	Clothing        string     `json:"clothing"`
	SpecialFeatures string     `json:"special_features"`
	MentalStatus    string     `json:"mental_status"`
	PhysicalStatus  string     `json:"physical_status"`
	MissingTime     time.Time  `json:"missing_time" binding:"required"`
	MissingLocation string     `json:"missing_location" binding:"required"`
	MissingLongitude float64   `json:"missing_longitude"`
	MissingLatitude  float64   `json:"missing_latitude"`
	MissingDetail   string     `json:"missing_detail"`
	PossibleLocation string    `json:"possible_location"`
	Photos          []string   `json:"photos"`
	ContactName     string     `json:"contact_name"`
	ContactPhone    string     `json:"contact_phone" binding:"required"`
	ContactRelation string     `json:"contact_relation"`
	ContactAddress  string     `json:"contact_address"`
	FamilyDescription string  `json:"family_description"`
	CaseType        string     `json:"case_type"`
	DialectIDs      []string   `json:"dialect_ids"`
	OrgID           uuid.UUID  `json:"org_id"`
}

// Create 创建走失人员记录
func (s *MissingPersonService) Create(ctx context.Context, req *CreateMissingPersonRequest, reporterID uuid.UUID) (*model.MissingPerson, error) {
	mp := &model.MissingPerson{
		Name:             req.Name,
		Gender:           req.Gender,
		BirthDate:        req.BirthDate,
		Age:              req.Age,
		Height:           req.Height,
		Weight:           req.Weight,
		IDCard:           req.IDCard,
		Appearance:       req.Appearance,
		Clothing:         req.Clothing,
		SpecialFeatures:  req.SpecialFeatures,
		MentalStatus:     req.MentalStatus,
		PhysicalStatus:   req.PhysicalStatus,
		MissingTime:      req.MissingTime,
		MissingLocation:  req.MissingLocation,
		MissingLongitude: req.MissingLongitude,
		MissingLatitude:  req.MissingLatitude,
		MissingDetail:    req.MissingDetail,
		PossibleLocation: req.PossibleLocation,
		ContactName:      req.ContactName,
		ContactPhone:     req.ContactPhone,
		ContactRelation:  req.ContactRelation,
		ContactAddress:   req.ContactAddress,
		FamilyDescription: req.FamilyDescription,
		CaseType:         req.CaseType,
		ReporterID:       reporterID,
		OrgID:            req.OrgID,
		Status:           model.CaseStatusMissing,
	}

	if err := s.mpRepo.Create(ctx, mp); err != nil {
		return nil, err
	}

	// 添加照片
	for i, url := range req.Photos {
		photo := &model.MissingPhoto{
			MissingPersonID: mp.ID,
			URL:             url,
			Type:            "normal",
			Sort:            i,
		}
		s.mpRepo.AddPhoto(ctx, photo)
	}

	// 更新组织案件数
	s.orgRepo.UpdateVolunteerCount(ctx, req.OrgID)

	return mp, nil
}

// GetByID 根据ID获取
func (s *MissingPersonService) GetByID(ctx context.Context, id uuid.UUID) (*model.MissingPerson, error) {
	// 增加浏览次数
	s.mpRepo.IncrementViewCount(ctx, id)
	return s.mpRepo.GetByID(ctx, id)
}

// UpdateStatus 更新状态
func (s *MissingPersonService) UpdateStatus(ctx context.Context, id uuid.UUID, status string, foundDetail string) error {
	mp, err := s.mpRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if status == model.CaseStatusFound || status == model.CaseStatusReunited {
		now := time.Now()
		mp.FoundTime = &now
		if foundDetail != "" {
			mp.FoundDetail = foundDetail
		}
	}

	mp.Status = status
	return s.mpRepo.Update(ctx, mp)
}

// List 列表查询
func (s *MissingPersonService) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.MissingPerson, int64, error) {
	return s.mpRepo.List(ctx, page, pageSize, filters)
}

// GetNearbyCases 获取附近案件
func (s *MissingPersonService) GetNearbyCases(ctx context.Context, lat, lng, radius float64) ([]*model.MissingPerson, error) {
	return s.mpRepo.GetNearbyCases(ctx, lat, lng, radius)
}

// AddTrack 添加轨迹
func (s *MissingPersonService) AddTrack(ctx context.Context, missingPersonID uuid.UUID, trackTime time.Time, location string, longitude, latitude float64, description string, photos []string, reporterID uuid.UUID) error {
	// 检查案件是否存在
	_, err := s.mpRepo.GetByID(ctx, missingPersonID)
	if err != nil {
		return fmt.Errorf("案件不存在")
	}

	track := &model.MissingPersonTrack{
		MissingPersonID: missingPersonID,
		ReporterID:      reporterID,
		TrackTime:       trackTime,
		Location:        location,
		Longitude:       longitude,
		Latitude:        latitude,
		Description:     description,
		Photos:          photos,
	}

	return s.mpRepo.AddTrack(ctx, track)
}

// GetTracks 获取轨迹列表
func (s *MissingPersonService) GetTracks(ctx context.Context, missingPersonID uuid.UUID) ([]*model.MissingPersonTrack, error) {
	return s.mpRepo.GetTracks(ctx, missingPersonID)
}

// GetStatistics 获取统计
func (s *MissingPersonService) GetStatistics(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	filters := map[string]interface{}{}
	if orgID != uuid.Nil {
		filters["org_id"] = orgID
	}

	// 总案件数
	total, _, err := s.mpRepo.List(ctx, 1, 1, filters)
	if err != nil {
		return nil, err
	}
	result["total"] = len(total)

	// 各状态统计
	statuses := []string{model.CaseStatusMissing, model.CaseStatusSearching, model.CaseStatusFound, model.CaseStatusReunited, model.CaseStatusClosed}
	for _, status := range statuses {
		filters["status"] = status
		list, _, err := s.mpRepo.List(ctx, 1, 1, filters)
		if err != nil {
			return nil, err
		}
		result[status] = len(list)
	}

	return result, nil
}
