package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateDialectRequest 创建方言请求
type CreateDialectRequest struct {
	Title            string  `json:"title" binding:"required"`
	Content          string  `json:"content"`
	Region           string  `json:"region" binding:"required"`
	Province         string  `json:"province"`
	City             string  `json:"city"`
	DialectType      string  `json:"dialect_type"`
	AudioUrl         string  `json:"audio_url" binding:"required"`
	Duration         int     `json:"duration"`
	FileSize         int     `json:"file_size"`
	Format           string  `json:"format"`
	Tags             string  `json:"tags"`
	Description      string  `json:"description"`
	MissingPersonID  string  `json:"missing_person_id"`
	CollectAddress   string  `json:"collect_address"`
	CollectLatitude  float64 `json:"collect_latitude"`
	CollectLongitude float64 `json:"collect_longitude"`
}

// CreateDialectBatchRequest 批量创建方言（按卡片分段）
type CreateDialectBatchRequest struct {
	Region           string                     `json:"region" binding:"required"`
	Province         string                     `json:"province"`
	City             string                     `json:"city"`
	District         string                     `json:"district"`
	Description      string                     `json:"description"`
	Tags             string                     `json:"tags"`
	MissingPersonID  string                     `json:"missing_person_id"`
	CollectAddress   string                     `json:"collect_address" binding:"required"`
	CollectLatitude  float64                    `json:"collect_latitude"`
	CollectLongitude float64                    `json:"collect_longitude"`
	Recordings       []DialectCardRecordingItem `json:"recordings" binding:"required,min=1,dive"`
}

// DialectCardRecordingItem 单个卡片录音项
type DialectCardRecordingItem struct {
	CardID   string `json:"card_id" binding:"required"`
	AudioURL string `json:"audio_url" binding:"required"`
	Duration int    `json:"duration" binding:"required,min=1"`
	FileSize int    `json:"file_size"`
	Format   string `json:"format"`
}

// UpdateDialectRequest 更新方言请求
type UpdateDialectRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Region      string `json:"region"`
	Province    string `json:"province"`
	City        string `json:"city"`
	DialectType string `json:"dialect_type"`
	Tags        string `json:"tags"`
	Description string `json:"description"`
}

// DialectResponse 方言响应
type DialectResponse struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Content          string        `json:"content"`
	Region           string        `json:"region"`
	Province         string        `json:"province"`
	City             string        `json:"city"`
	DialectType      string        `json:"dialect_type"`
	AudioUrl         string        `json:"audio_url"`
	Duration         int           `json:"duration"`
	FileSize         int           `json:"file_size"`
	Format           string        `json:"format"`
	Status           string        `json:"status"`
	IsFeatured       bool          `json:"is_featured"`
	PlayCount        int           `json:"play_count"`
	LikeCount        int           `json:"like_count"`
	CommentCount     int           `json:"comment_count"`
	Tags             string        `json:"tags"`
	Description      string        `json:"description"`
	CollectAddress   string        `json:"collect_address"`
	CollectLatitude  float64       `json:"collect_latitude"`
	CollectLongitude float64       `json:"collect_longitude"`
	BatchID          *string       `json:"batch_id,omitempty"`
	CardGroupID      *string       `json:"card_group_id,omitempty"`
	CardID           *string       `json:"card_id,omitempty"`
	CardContent      string        `json:"card_content,omitempty"`
	CardImageURL     string        `json:"card_image_url,omitempty"`
	UploaderID       string        `json:"uploader_id"`
	OrgID            string        `json:"org_id"`
	MissingPersonID  *string       `json:"missing_person_id,omitempty"`
	IsLiked          bool          `json:"is_liked"`
	Uploader         *UserResponse `json:"uploader,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
}

// DialectCardGroupResponse 方言卡片分组响应
type DialectCardGroupResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	SortOrder   int                   `json:"sort_order"`
	Status      string                `json:"status"`
	Cards       []DialectCardResponse `json:"cards,omitempty"`
}

// DialectCardResponse 方言卡片响应
type DialectCardResponse struct {
	ID        string `json:"id"`
	GroupID   string `json:"group_id"`
	Content   string `json:"content"`
	ImageURL  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
	Required  bool   `json:"required"`
	Status    string `json:"status"`
}

// CreateDialectCardGroupRequest 创建卡片分组
type CreateDialectCardGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status"`
}

// UpdateDialectCardGroupRequest 更新卡片分组
type UpdateDialectCardGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   *int   `json:"sort_order"`
	Status      string `json:"status"`
}

// CreateDialectCardRequest 创建卡片
type CreateDialectCardRequest struct {
	GroupID   string `json:"group_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
	ImageURL  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
	Required  *bool  `json:"required"`
	Status    string `json:"status"`
}

// UpdateDialectCardRequest 更新卡片
type UpdateDialectCardRequest struct {
	GroupID   string `json:"group_id"`
	Content   string `json:"content"`
	ImageURL  string `json:"image_url"`
	SortOrder *int   `json:"sort_order"`
	Required  *bool  `json:"required"`
	Status    string `json:"status"`
}

// DialectBatchCreateResponse 批量录入响应
type DialectBatchCreateResponse struct {
	BatchID string            `json:"batch_id"`
	Total   int               `json:"total"`
	Items   []DialectResponse `json:"items"`
}

// DialectListRequest 方言列表请求
type DialectListRequest struct {
	Page      int    `form:"page,default=1" binding:"min=1"`
	PageSize  int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword   string `form:"keyword"`
	Region    string `form:"region"`
	Province  string `form:"province"`
	City      string `form:"city"`
	Type      string `form:"type"`
	Status    string `form:"status"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// DialectListResponse 方言列表响应
type DialectListResponse = PageResult[DialectResponse]

// DialectCommentListResponse 方言评论列表响应
type DialectCommentListResponse = PageResult[DialectCommentResponse]

// UpdateDialectStatusRequest 更新状态请求
type UpdateDialectStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// CreateDialectCommentRequest 创建评论请求
type CreateDialectCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID string `json:"parent_id"`
}

// DialectCommentResponse 评论响应
type DialectCommentResponse struct {
	ID         string        `json:"id"`
	DialectID  string        `json:"dialect_id"`
	UserID     string        `json:"user_id"`
	Content    string        `json:"content"`
	ParentID   *string       `json:"parent_id,omitempty"`
	ReplyCount int           `json:"reply_count"`
	LikeCount  int           `json:"like_count"`
	User       *UserResponse `json:"user,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

// DialectStatsResponse 方言统计响应
type DialectStatsResponse struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Pending    int64 `json:"pending"`
	Featured   int64 `json:"featured"`
	TotalPlays int64 `json:"total_plays"`
	TotalLikes int64 `json:"total_likes"`
}

// ToDialectResponse 转换为方言响应
func ToDialectResponse(d *entity.Dialect, isLiked bool) DialectResponse {
	resp := DialectResponse{
		ID:               d.ID,
		Title:            d.Title,
		Content:          d.Content,
		Region:           d.Region,
		Province:         d.Province,
		City:             d.City,
		DialectType:      string(d.DialectType),
		AudioUrl:         d.AudioUrl,
		Duration:         d.Duration,
		FileSize:         d.FileSize,
		Format:           d.Format,
		Status:           string(d.Status),
		IsFeatured:       d.IsFeatured,
		PlayCount:        d.PlayCount,
		LikeCount:        d.LikeCount,
		CommentCount:     d.CommentCount,
		Tags:             d.Tags,
		Description:      d.Description,
		CollectAddress:   d.CollectAddress,
		CollectLatitude:  d.CollectLatitude,
		CollectLongitude: d.CollectLongitude,
		BatchID:          d.BatchID,
		CardGroupID:      d.CardGroupID,
		CardID:           d.CardID,
		UploaderID:       d.UploaderID,
		OrgID:            d.OrgID,
		MissingPersonID:  d.MissingPersonID,
		IsLiked:          isLiked,
		CreatedAt:        d.CreatedAt,
	}

	if d.Card != nil {
		resp.CardContent = d.Card.Content
		resp.CardImageURL = d.Card.ImageURL
	}

	if d.Uploader != nil {
		uploader := ToUserResponse(d.Uploader)
		resp.Uploader = &uploader
	}

	return resp
}

func ToDialectCardGroupResponse(group *entity.DialectCardGroup) DialectCardGroupResponse {
	resp := DialectCardGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		SortOrder:   group.SortOrder,
		Status:      string(group.Status),
	}
	if len(group.Cards) > 0 {
		resp.Cards = make([]DialectCardResponse, 0, len(group.Cards))
		for idx := range group.Cards {
			resp.Cards = append(resp.Cards, ToDialectCardResponse(&group.Cards[idx]))
		}
	}
	return resp
}

func ToDialectCardResponse(card *entity.DialectCard) DialectCardResponse {
	return DialectCardResponse{
		ID:        card.ID,
		GroupID:   card.GroupID,
		Content:   card.Content,
		ImageURL:  card.ImageURL,
		SortOrder: card.SortOrder,
		Required:  card.Required,
		Status:    string(card.Status),
	}
}

// ToDialectCommentResponse 转换为评论响应
func ToDialectCommentResponse(c *entity.DialectComment) DialectCommentResponse {
	resp := DialectCommentResponse{
		ID:         c.ID,
		DialectID:  c.DialectID,
		UserID:     c.UserID,
		Content:    c.Content,
		ParentID:   c.ParentID,
		ReplyCount: c.ReplyCount,
		LikeCount:  c.LikeCount,
		CreatedAt:  c.CreatedAt,
	}

	if c.User != nil {
		user := ToUserResponse(c.User)
		resp.User = &user
	}

	return resp
}

// NewDialectListResponse 创建方言列表响应
func NewDialectListResponse(list []DialectResponse, total int64, page, pageSize int) DialectListResponse {
	if pageSize <= 0 {
		pageSize = 10
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return DialectListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
