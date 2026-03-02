package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateDialectRequest 创建方言请求
type CreateDialectRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content"`
	Region      string `json:"region" binding:"required"`
	Province    string `json:"province"`
	City        string `json:"city"`
	DialectType string `json:"dialect_type"`
	AudioUrl    string `json:"audio_url" binding:"required"`
	Duration    int    `json:"duration"`
	FileSize    int    `json:"file_size"`
	Format      string `json:"format"`
	Tags        string `json:"tags"`
	Description string `json:"description"`
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
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Region       string    `json:"region"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	DialectType  string    `json:"dialect_type"`
	AudioUrl     string    `json:"audio_url"`
	Duration     int       `json:"duration"`
	FileSize     int       `json:"file_size"`
	Format       string    `json:"format"`
	Status       string    `json:"status"`
	IsFeatured   bool      `json:"is_featured"`
	PlayCount    int       `json:"play_count"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	Tags         string    `json:"tags"`
	Description  string    `json:"description"`
	UploaderID   string    `json:"uploader_id"`
	OrgID        string    `json:"org_id"`
	Uploader     *UserResponse `json:"uploader,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// DialectListRequest 方言列表请求
type DialectListRequest struct {
	Page       int    `form:"page,default=1" binding:"min=1"`
	PageSize   int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword    string `form:"keyword"`
	Region     string `form:"region"`
	Province   string `form:"province"`
	City       string `form:"city"`
	Type       string `form:"type"`
	Status     string `form:"status"`
	SortBy     string `form:"sort_by"`
	SortOrder  string `form:"sort_order"`
}

// DialectListResponse 方言列表响应
type DialectListResponse = PageResult[DialectResponse]

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
	ID        string    `json:"id"`
	DialectID string    `json:"dialect_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	ParentID  *string   `json:"parent_id,omitempty"`
	ReplyCount int      `json:"reply_count"`
	LikeCount int       `json:"like_count"`
	User      *UserResponse `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// DialectStatsResponse 方言统计响应
type DialectStatsResponse struct {
	Total       int64 `json:"total"`
	Active      int64 `json:"active"`
	Pending     int64 `json:"pending"`
	Featured    int64 `json:"featured"`
	TotalPlays  int64 `json:"total_plays"`
	TotalLikes  int64 `json:"total_likes"`
}

// ToDialectResponse 转换为方言响应
func ToDialectResponse(d *entity.Dialect) DialectResponse {
	resp := DialectResponse{
		ID:           d.ID,
		Title:        d.Title,
		Content:      d.Content,
		Region:       d.Region,
		Province:     d.Province,
		City:         d.City,
		DialectType:  string(d.DialectType),
		AudioUrl:     d.AudioUrl,
		Duration:     d.Duration,
		FileSize:     d.FileSize,
		Format:       d.Format,
		Status:       string(d.Status),
		IsFeatured:   d.IsFeatured,
		PlayCount:    d.PlayCount,
		LikeCount:    d.LikeCount,
		CommentCount: d.CommentCount,
		Tags:         d.Tags,
		Description:  d.Description,
		UploaderID:   d.UploaderID,
		OrgID:        d.OrgID,
		CreatedAt:    d.CreatedAt,
	}

	if d.Uploader != nil {
		uploader := ToUserResponse(d.Uploader)
		resp.Uploader = &uploader
	}

	return resp
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
