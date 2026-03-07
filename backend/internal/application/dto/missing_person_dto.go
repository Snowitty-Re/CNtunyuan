package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateMissingPersonRequest 创建走失人员请求
type CreateMissingPersonRequest struct {
	Name         string    `json:"name" binding:"required"`
	Gender       string    `json:"gender" binding:"required"`
	BirthDate    time.Time `json:"birth_date"`
	Age          int       `json:"age"`
	Height       int       `json:"height"`
	Weight       int       `json:"weight"`
	Description  string    `json:"description"`
	PhotoUrl     string    `json:"photo_url"`
	MissingTime  time.Time `json:"missing_time" binding:"required"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	District     string    `json:"district"`
	Address      string    `json:"address"`
	Clothes      string    `json:"clothes"`
	Features     string    `json:"features"`
	ContactName  string    `json:"contact_name" binding:"required"`
	ContactPhone string    `json:"contact_phone" binding:"required"`
	ContactRel   string    `json:"contact_rel"`
	AltContact   string    `json:"alt_contact"`
	UrgencyLevel string    `json:"urgency_level"`
}

// UpdateMissingPersonRequest 更新走失人员请求
type UpdateMissingPersonRequest struct {
	Name         string    `json:"name"`
	Gender       string    `json:"gender"`
	BirthDate    time.Time `json:"birth_date"`
	Age          int       `json:"age"`
	Height       int       `json:"height"`
	Weight       int       `json:"weight"`
	Description  string    `json:"description"`
	PhotoUrl     string    `json:"photo_url"`
	MissingTime  time.Time `json:"missing_time"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	District     string    `json:"district"`
	Address      string    `json:"address"`
	Clothes      string    `json:"clothes"`
	Features     string    `json:"features"`
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	ContactRel   string    `json:"contact_rel"`
	AltContact   string    `json:"alt_contact"`
	UrgencyLevel string    `json:"urgency_level"`
}

// MissingPersonPhoto 走失人员照片响应
type MissingPersonPhoto struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Type      string `json:"type"`
	IsPrimary bool   `json:"is_primary"`
}

// MissingPersonResponse 走失人员响应
type MissingPersonResponse struct {
	ID            string               `json:"id"`
	CaseNo        string               `json:"case_no"`
	Name          string               `json:"name"`
	Gender        string               `json:"gender"`
	BirthDate     *time.Time           `json:"birth_date,omitempty"`
	Age           int                  `json:"age"`
	Height        int                  `json:"height"`
	Weight        int                  `json:"weight"`
	Description   string               `json:"description"`
	PhotoUrl      string               `json:"photo_url"`
	MissingTime   time.Time            `json:"missing_time"`
	Province      string               `json:"province"`
	City          string               `json:"city"`
	District      string               `json:"district"`
	Address       string               `json:"address"`
	Clothes       string               `json:"clothes"`
	Features      string               `json:"features"`
	ContactName   string               `json:"contact_name"`
	ContactPhone  string               `json:"contact_phone"`
	ContactRel    string               `json:"contact_rel"`
	AltContact    string               `json:"alt_contact"`
	Status        string               `json:"status"`
	Urgency       string               `json:"urgency"`
	Views         int                  `json:"views"`
	ShareCount    int                  `json:"share_count"`
	ReporterID    string               `json:"reporter_id"`
	OrgID         string               `json:"org_id"`
	AssignedTo    *string              `json:"assigned_to,omitempty"`
	FoundTime     *time.Time           `json:"found_time,omitempty"`
	FoundLocation string               `json:"found_location"`
	FoundNote     string               `json:"found_note"`
	Reporter      *UserResponse        `json:"reporter,omitempty"`
	Assignee      *UserResponse        `json:"assignee,omitempty"`
	Photos        []MissingPersonPhoto `json:"photos,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
}

// MissingPersonListRequest 走失人员列表请求
type MissingPersonListRequest struct {
	Page         int    `form:"page,default=1" binding:"min=1"`
	PageSize     int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword      string `form:"keyword"`
	Status       string `form:"status"`
	Gender       string `form:"gender"`
	AgeMin       int    `form:"age_min"`
	AgeMax       int    `form:"age_max"`
	Province     string `form:"province"`
	City         string `form:"city"`
	District     string `form:"district"`
	UrgencyLevel string `form:"urgency_level"`
}

// MissingPersonListResponse 走失人员列表响应
type MissingPersonListResponse = PageResult[MissingPersonResponse]

// UpdateMissingPersonStatusRequest 更新状态请求
type UpdateMissingPersonStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// MarkFoundRequest 标记找到请求
type MarkFoundRequest struct {
	Location string `json:"location"`
	Note     string `json:"note"`
}

// CreateMissingPersonTrackRequest 创建轨迹请求
type CreateMissingPersonTrackRequest struct {
	Location    string    `json:"location"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	District    string    `json:"district"`
	Address     string    `json:"address"`
	Time        time.Time `json:"time"`
	Description string    `json:"description"`
	IsKeyPoint  bool      `json:"is_key_point"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
}

// MissingPersonTrackResponse 轨迹响应
type MissingPersonTrackResponse struct {
	ID              string        `json:"id"`
	MissingPersonID string        `json:"missing_person_id"`
	ReporterID      string        `json:"reporter_id"`
	Location        string        `json:"location"`
	Province        string        `json:"province"`
	City            string        `json:"city"`
	District        string        `json:"district"`
	Address         string        `json:"address"`
	Time            time.Time     `json:"time"`
	Description     string        `json:"description"`
	IsKeyPoint      bool          `json:"is_key_point"`
	Lat             float64       `json:"lat"`
	Lng             float64       `json:"lng"`
	Status          string        `json:"status"`
	Reporter        *UserResponse `json:"reporter,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

// MissingPersonStatsResponse 统计响应
type MissingPersonStatsResponse struct {
	Total     int64 `json:"total"`
	Missing   int64 `json:"missing"`
	Searching int64 `json:"searching"`
	Found     int64 `json:"found"`
	Reunited  int64 `json:"reunited"`
	Closed    int64 `json:"closed"`
	TodayNew  int64 `json:"today_new"`
	WeekNew   int64 `json:"week_new"`
	MonthNew  int64 `json:"month_new"`
}

// ToMissingPersonResponse 转换为走失人员响应
func ToMissingPersonResponse(mp *entity.MissingPerson) MissingPersonResponse {
	resp := MissingPersonResponse{
		ID:            mp.ID,
		CaseNo:        mp.CaseNo,
		Name:          mp.Name,
		Gender:        mp.Gender,
		BirthDate:     mp.BirthDate,
		Age:           mp.Age,
		Height:        mp.Height,
		Weight:        mp.Weight,
		Description:   mp.Description,
		PhotoUrl:      mp.PhotoUrl,
		MissingTime:   mp.MissingTime,
		Province:      mp.Province,
		City:          mp.City,
		District:      mp.District,
		Address:       mp.Address,
		Clothes:       mp.Clothes,
		Features:      mp.Features,
		ContactName:   mp.ContactName,
		ContactPhone:  mp.ContactPhone,
		ContactRel:    mp.ContactRel,
		AltContact:    mp.AltContact,
		Status:        string(mp.Status),
		Urgency:       string(mp.Urgency),
		Views:         mp.Views,
		ShareCount:    mp.ShareCount,
		ReporterID:    mp.ReporterID,
		OrgID:         mp.OrgID,
		AssignedTo:    mp.AssignedTo,
		FoundTime:     mp.FoundTime,
		FoundLocation: mp.FoundLocation,
		FoundNote:     mp.FoundNote,
		CreatedAt:     mp.CreatedAt,
	}

	if mp.Reporter != nil {
		reporter := ToUserResponse(mp.Reporter)
		resp.Reporter = &reporter
	}
	if mp.Assignee != nil {
		assignee := ToUserResponse(mp.Assignee)
		resp.Assignee = &assignee
	}

	// 转换照片列表
	if len(mp.Photos) > 0 {
		resp.Photos = make([]MissingPersonPhoto, len(mp.Photos))
		for i, photo := range mp.Photos {
			resp.Photos[i] = MissingPersonPhoto{
				ID:        photo.ID,
				URL:       photo.URL,
				Type:      photo.Type,
				IsPrimary: photo.IsPrimary,
			}
		}
	}

	return resp
}

// ToMissingPersonTrackResponse 转换为轨迹响应
func ToMissingPersonTrackResponse(track *entity.MissingPersonTrack) MissingPersonTrackResponse {
	resp := MissingPersonTrackResponse{
		ID:              track.ID,
		MissingPersonID: track.MissingPersonID,
		ReporterID:      track.ReporterID,
		Location:        track.Location,
		Province:        track.Province,
		City:            track.City,
		District:        track.District,
		Address:         track.Address,
		Time:            track.Time,
		Description:     track.Description,
		IsKeyPoint:      track.IsKeyPoint,
		Lat:             track.Lat,
		Lng:             track.Lng,
		Status:          track.Status,
		CreatedAt:       track.CreatedAt,
	}

	if track.Reporter != nil {
		reporter := ToUserResponse(track.Reporter)
		resp.Reporter = &reporter
	}

	return resp
}

// NewMissingPersonListResponse 创建走失人员列表响应
func NewMissingPersonListResponse(list []MissingPersonResponse, total int64, page, pageSize int) MissingPersonListResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return MissingPersonListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
