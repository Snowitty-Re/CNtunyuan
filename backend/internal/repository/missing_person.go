package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MissingPersonRepository 走失人员仓库
type MissingPersonRepository struct {
	db *gorm.DB
}

// NewMissingPersonRepository 创建仓库
func NewMissingPersonRepository(db *gorm.DB) *MissingPersonRepository {
	return &MissingPersonRepository{db: db}
}

// Create 创建
func (r *MissingPersonRepository) Create(ctx context.Context, mp *model.MissingPerson) error {
	return r.db.WithContext(ctx).Create(mp).Error
}

// GetByID 根据ID获取
func (r *MissingPersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MissingPerson, error) {
	var mp model.MissingPerson
	err := r.db.WithContext(ctx).Preload("Photos").Preload("Dialects").Preload("Reporter").Preload("Org").First(&mp, id).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// GetByCaseNo 根据案件编号获取
func (r *MissingPersonRepository) GetByCaseNo(ctx context.Context, caseNo string) (*model.MissingPerson, error) {
	var mp model.MissingPerson
	err := r.db.WithContext(ctx).Where("case_no = ?", caseNo).First(&mp).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// Update 更新
func (r *MissingPersonRepository) Update(ctx context.Context, mp *model.MissingPerson) error {
	return r.db.WithContext(ctx).Save(mp).Error
}

// Delete 删除
func (r *MissingPersonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MissingPerson{}, id).Error
}

// List 列表查询
func (r *MissingPersonRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*model.MissingPerson, int64, error) {
	var mps []*model.MissingPerson
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MissingPerson{})

	// 过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "name":
				query = query.Where("name LIKE ?", "%"+value.(string)+"%")
			case "status":
				query = query.Where("status = ?", value)
			case "case_type":
				query = query.Where("case_type = ?", value)
			case "org_id":
				query = query.Where("org_id = ?", value)
			case "province":
				query = query.Where("missing_location LIKE ?", "%"+value.(string)+"%")
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Photos").Preload("Reporter").Preload("Org").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&mps).Error; err != nil {
		return nil, 0, err
	}

	return mps, total, nil
}

// UpdateStatus 更新状态
func (r *MissingPersonRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.MissingPerson{}).Where("id = ?", id).Update("status", status).Error
}

// IncrementViewCount 增加浏览次数
func (r *MissingPersonRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.MissingPerson{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GetNearbyCases 获取附近案件
func (r *MissingPersonRepository) GetNearbyCases(ctx context.Context, lat, lng float64, radius float64) ([]*model.MissingPerson, error) {
	var mps []*model.MissingPerson
	
	// 使用PostGIS或简化计算
	// 这里使用简单的距离公式
	err := r.db.WithContext(ctx).Where(
		"missing_latitude BETWEEN ? AND ? AND missing_longitude BETWEEN ? AND ?",
		lat-radius, lat+radius, lng-radius, lng+radius,
	).Where("status IN ?", []string{model.CaseStatusMissing, model.CaseStatusSearching}).
		Find(&mps).Error
	
	return mps, err
}

// AddPhoto 添加照片
func (r *MissingPersonRepository) AddPhoto(ctx context.Context, photo *model.MissingPhoto) error {
	return r.db.WithContext(ctx).Create(photo).Error
}

// DeletePhoto 删除照片
func (r *MissingPersonRepository) DeletePhoto(ctx context.Context, photoID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MissingPhoto{}, photoID).Error
}

// AddTrack 添加轨迹
func (r *MissingPersonRepository) AddTrack(ctx context.Context, track *model.MissingPersonTrack) error {
	return r.db.WithContext(ctx).Create(track).Error
}

// GetTracks 获取轨迹列表
func (r *MissingPersonRepository) GetTracks(ctx context.Context, missingPersonID uuid.UUID) ([]*model.MissingPersonTrack, error) {
	var tracks []*model.MissingPersonTrack
	err := r.db.WithContext(ctx).Where("missing_person_id = ?", missingPersonID).Preload("Reporter").Order("track_time DESC").Find(&tracks).Error
	return tracks, err
}
