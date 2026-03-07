package dto

import (
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code" binding:"required"`
	Type         string `json:"type" binding:"required"`
	ParentID     string `json:"parent_id"`
	Description  string `json:"description"`
	Address      string `json:"address"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	Description  string `json:"description"`
	Address      string `json:"address"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	Status       string `json:"status"`
	SortOrder    int    `json:"sort_order"`
}

// OrganizationResponse 组织响应
type OrganizationResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Code         string                 `json:"code"`
	Type         string                 `json:"type"`
	Level        int                    `json:"level"`
	ParentID     *string                `json:"parent_id,omitempty"`
	Description  string                 `json:"description"`
	Address      string                 `json:"address"`
	ContactName  string                 `json:"contact_name"`
	ContactPhone string                 `json:"contact_phone"`
	Status       string                 `json:"status"`
	SortOrder    int                    `json:"sort_order"`
	CreatedAt    time.Time              `json:"created_at"`
	Children     []OrganizationResponse `json:"children,omitempty"`
}

// OrganizationTreeResponse 组织树响应
type OrganizationTreeResponse struct {
	OrganizationResponse
	Children []*OrganizationTreeResponse `json:"children,omitempty"`
}

// OrganizationListRequest 组织列表请求
type OrganizationListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Status   string `form:"status"`
	ParentID string `form:"parent_id"`
}

// OrganizationListResponse 组织列表响应
type OrganizationListResponse = PageResult[OrganizationResponse]

// MoveOrganizationRequest 移动组织请求
type MoveOrganizationRequest struct {
	NewParentID string `json:"new_parent_id" binding:"required"`
}

// OrganizationStatsResponse 组织统计响应
type OrganizationStatsResponse struct {
	TotalVolunteers  int64 `json:"total_volunteers"`
	ActiveVolunteers int64 `json:"active_volunteers"`
	TotalCases       int64 `json:"total_cases"`
	ActiveCases      int64 `json:"active_cases"`
	CompletedCases   int64 `json:"completed_cases"`
	TotalTasks       int64 `json:"total_tasks"`
	PendingTasks     int64 `json:"pending_tasks"`
}

// ToOrganizationResponse 转换为组织响应
func ToOrganizationResponse(org *entity.Organization) OrganizationResponse {
	resp := OrganizationResponse{
		ID:           org.ID,
		Name:         org.Name,
		Code:         org.Code,
		Type:         string(org.Type),
		Level:        org.Level,
		ParentID:     org.ParentID,
		Description:  org.Description,
		Address:      org.Address,
		ContactName:  org.ContactName,
		ContactPhone: org.ContactPhone,
		Status:       string(org.Status),
		SortOrder:    org.SortOrder,
		CreatedAt:    org.CreatedAt,
	}

	if org.Children != nil {
		resp.Children = make([]OrganizationResponse, len(org.Children))
		for i, child := range org.Children {
			resp.Children[i] = ToOrganizationResponse(&child)
		}
	}

	return resp
}

// ToOrganizationTreeResponse 转换为组织树响应
func ToOrganizationTreeResponse(node *entity.OrgTreeNode) *OrganizationTreeResponse {
	if node == nil {
		return nil
	}

	resp := &OrganizationTreeResponse{
		OrganizationResponse: OrganizationResponse{
			ID:           node.ID,
			Name:         node.Name,
			Code:         node.Code,
			Type:         string(node.Type),
			Level:        node.Level,
			ParentID:     node.ParentID,
			Description:  node.Description,
			Address:      node.Address,
			ContactName:  node.ContactName,
			ContactPhone: node.ContactPhone,
			Status:       string(node.Status),
			SortOrder:    node.SortOrder,
			CreatedAt:    node.CreatedAt,
		},
	}

	if len(node.Children) > 0 {
		resp.Children = make([]*OrganizationTreeResponse, len(node.Children))
		for i, child := range node.Children {
			resp.Children[i] = ToOrganizationTreeResponse(child)
		}
	}

	return resp
}

// NewOrganizationListResponse 创建组织列表响应
func NewOrganizationListResponse(list []OrganizationResponse, total int64, page, pageSize int) OrganizationListResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return OrganizationListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
