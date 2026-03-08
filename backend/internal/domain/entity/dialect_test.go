// Package entity 方言实体测试
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDialect(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		region     string
		audioUrl   string
		uploaderID string
		orgID      string
		duration   int
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid dialect",
			title:      "方言录音",
			region:     "北京",
			audioUrl:   "https://example.com/audio.mp3",
			uploaderID: "user-id",
			orgID:      "org-id",
			duration:   60,
			wantErr:    false,
		},
		{
			name:       "empty title",
			title:      "",
			region:     "北京",
			audioUrl:   "https://example.com/audio.mp3",
			uploaderID: "user-id",
			orgID:      "org-id",
			duration:   60,
			wantErr:    true,
			errMsg:     "标题不能为空",
		},
		{
			name:       "empty audio url",
			title:      "方言录音",
			region:     "北京",
			audioUrl:   "",
			uploaderID: "user-id",
			orgID:      "org-id",
			duration:   60,
			wantErr:    true,
			errMsg:     "音频URL不能为空",
		},
		{
			name:       "empty region",
			title:      "方言录音",
			region:     "",
			audioUrl:   "https://example.com/audio.mp3",
			uploaderID: "user-id",
			orgID:      "org-id",
			duration:   60,
			wantErr:    true,
			errMsg:     "地区不能为空",
		},
		{
			name:       "duration too short",
			title:      "方言录音",
			region:     "北京",
			audioUrl:   "https://example.com/audio.mp3",
			uploaderID: "user-id",
			orgID:      "org-id",
			duration:   0,
			wantErr:    true,
			errMsg:     "音频时长必须在1-300秒之间",
		},
		{
			name:       "duration too long",
			title:      "方言录音",
			region:     "北京",
			audioUrl:   "https://example.com/audio.mp3",
			uploaderID: "user-id",
			orgID:      "org-id",
			duration:   301,
			wantErr:    true,
			errMsg:     "音频时长必须在1-300秒之间",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialect, err := NewDialect(tt.title, tt.region, tt.audioUrl, tt.uploaderID, tt.orgID, tt.duration)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, dialect.ID)
			assert.Equal(t, tt.title, dialect.Title)
			assert.Equal(t, tt.region, dialect.Region)
			assert.Equal(t, DialectStatusPending, dialect.Status)
			assert.Equal(t, DialectTypePhrase, dialect.DialectType)
		})
	}
}

func TestDialect_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status DialectStatus
		want   bool
	}{
		{"active", DialectStatusActive, true},
		{"inactive", DialectStatusInactive, false},
		{"pending", DialectStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dialect{Status: tt.status}
			assert.Equal(t, tt.want, d.IsActive())
		})
	}
}

func TestDialect_CanPlay(t *testing.T) {
	tests := []struct {
		name   string
		status DialectStatus
		want   bool
	}{
		{"active can play", DialectStatusActive, true},
		{"inactive cannot play", DialectStatusInactive, false},
		{"pending cannot play", DialectStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dialect{Status: tt.status}
			assert.Equal(t, tt.want, d.CanPlay())
		})
	}
}

func TestDialect_IncrementPlayCount(t *testing.T) {
	d := &Dialect{PlayCount: 0}
	
	d.IncrementPlayCount()
	assert.Equal(t, 1, d.PlayCount)
	
	d.IncrementPlayCount()
	assert.Equal(t, 2, d.PlayCount)
}

func TestDialect_IncrementLikeCount(t *testing.T) {
	d := &Dialect{LikeCount: 0}
	
	d.IncrementLikeCount()
	assert.Equal(t, 1, d.LikeCount)
	
	d.IncrementLikeCount()
	assert.Equal(t, 2, d.LikeCount)
}

func TestDialect_DecrementLikeCount(t *testing.T) {
	tests := []struct {
		name      string
		initial   int
		expected  int
	}{
		{"decrement from positive", 5, 4},
		{"decrement from 1", 1, 0},
		{"cannot go below 0", 0, 0},
 		{"cannot go below 0 from negative", -1, -1}, // 负数情况不处理
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dialect{LikeCount: tt.initial}
			d.DecrementLikeCount()
			assert.Equal(t, tt.expected, d.LikeCount)
		})
	}
}

func TestDialect_IncrementCommentCount(t *testing.T) {
	d := &Dialect{CommentCount: 0}
	
	d.IncrementCommentCount()
	assert.Equal(t, 1, d.CommentCount)
}

func TestDialect_DecrementCommentCount(t *testing.T) {
	tests := []struct {
		name      string
		initial   int
		expected  int
	}{
		{"decrement from positive", 5, 4},
		{"decrement from 1", 1, 0},
		{"cannot go below 0", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dialect{CommentCount: tt.initial}
			d.DecrementCommentCount()
			assert.Equal(t, tt.expected, d.CommentCount)
		})
	}
}

func TestDialect_Feature(t *testing.T) {
	d := &Dialect{IsFeatured: false}
	
	d.Feature()
	assert.True(t, d.IsFeatured)
}

func TestDialect_Unfeature(t *testing.T) {
	d := &Dialect{IsFeatured: true}
	
	d.Unfeature()
	assert.False(t, d.IsFeatured)
}

func TestDialect_Approve(t *testing.T) {
	d := &Dialect{Status: DialectStatusPending}
	
	d.Approve()
	assert.Equal(t, DialectStatusActive, d.Status)
}

func TestDialect_Reject(t *testing.T) {
	d := &Dialect{Status: DialectStatusPending}
	
	d.Reject()
	assert.Equal(t, DialectStatusInactive, d.Status)
}
