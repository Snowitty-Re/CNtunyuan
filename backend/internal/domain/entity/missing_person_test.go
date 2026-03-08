// Package entity 走失人员实体测试
package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMissingPerson(t *testing.T) {
	tests := []struct {
		name         string
		personName   string
		gender       string
		contactName  string
		contactPhone string
		reporterID   string
		orgID        string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid missing person",
			personName:   "张三",
			gender:       "男",
			contactName:  "李四",
			contactPhone: "13800138000",
			reporterID:   "reporter-id",
			orgID:        "org-id",
			wantErr:      false,
		},
		{
			name:         "empty name",
			personName:   "",
			gender:       "男",
			contactName:  "李四",
			contactPhone: "13800138000",
			reporterID:   "reporter-id",
			orgID:        "org-id",
			wantErr:      true,
			errMsg:       "姓名不能为空",
		},
		{
			name:         "empty gender",
			personName:   "张三",
			gender:       "",
			contactName:  "李四",
			contactPhone: "13800138000",
			reporterID:   "reporter-id",
			orgID:        "org-id",
			wantErr:      true,
			errMsg:       "性别不能为空",
		},
		{
			name:         "empty contact name",
			personName:   "张三",
			gender:       "男",
			contactName:  "",
			contactPhone: "13800138000",
			reporterID:   "reporter-id",
			orgID:        "org-id",
			wantErr:      true,
			errMsg:       "联系人姓名不能为空",
		},
		{
			name:         "empty contact phone",
			personName:   "张三",
			gender:       "男",
			contactName:  "李四",
			contactPhone: "",
			reporterID:   "reporter-id",
			orgID:        "org-id",
			wantErr:      true,
			errMsg:       "联系人电话不能为空",
		},
		// Note: MissingTime is automatically set to current time in NewMissingPerson,
		// so "zero missing time" validation test is no longer applicable here.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMissingPerson(tt.personName, tt.gender, tt.contactName, tt.contactPhone, tt.reporterID, tt.orgID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, mp.ID)
			assert.NotEmpty(t, mp.CaseNo)
			assert.Equal(t, tt.personName, mp.Name)
			assert.Equal(t, tt.gender, mp.Gender)
			assert.Equal(t, MissingStatusMissing, mp.Status)
			assert.Equal(t, UrgencyLevelMedium, mp.Urgency)
		})
	}
}

func TestMissingPerson_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status MissingStatus
		want   bool
	}{
		{"missing", MissingStatusMissing, true},
		{"searching", MissingStatusSearching, true},
		{"found", MissingStatusFound, false},
		{"reunited", MissingStatusReunited, false},
		{"closed", MissingStatusClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MissingPerson{Status: tt.status}
			assert.Equal(t, tt.want, mp.IsActive())
		})
	}
}

func TestMissingPerson_IsFound(t *testing.T) {
	tests := []struct {
		name   string
		status MissingStatus
		want   bool
	}{
		{"missing", MissingStatusMissing, false},
		{"searching", MissingStatusSearching, false},
		{"found", MissingStatusFound, true},
		{"reunited", MissingStatusReunited, true},
		{"closed", MissingStatusClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MissingPerson{Status: tt.status}
			assert.Equal(t, tt.want, mp.IsFound())
		})
	}
}

func TestMissingPerson_StartSearch(t *testing.T) {
	tests := []struct {
		name      string
		status    MissingStatus
		wantErr   bool
		errMsg    string
		newStatus MissingStatus
	}{
		{
			name:      "from missing to searching",
			status:    MissingStatusMissing,
			wantErr:   false,
			newStatus: MissingStatusSearching,
		},
		{
			name:    "from searching cannot start",
			status:  MissingStatusSearching,
			wantErr: true,
			errMsg:  "只有待寻找状态才能开始搜索",
		},
		{
			name:    "from found cannot start",
			status:  MissingStatusFound,
			wantErr: true,
			errMsg:  "只有待寻找状态才能开始搜索",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MissingPerson{Status: tt.status}
			err := mp.StartSearch()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.newStatus, mp.Status)
		})
	}
}

func TestMissingPerson_MarkFound(t *testing.T) {
	mp := &MissingPerson{
		Status: MissingStatusSearching,
	}
	
	location := "北京朝阳区"
	note := "在公园找到"
	
	err := mp.MarkFound(location, note)
	require.NoError(t, err)
	
	assert.Equal(t, MissingStatusFound, mp.Status)
	assert.Equal(t, location, mp.FoundLocation)
	assert.Equal(t, note, mp.FoundNote)
	assert.NotNil(t, mp.FoundTime)
	
	// Cannot mark found again
	err = mp.MarkFound(location, note)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该案件已被标记为找到")
}

func TestMissingPerson_MarkReunited(t *testing.T) {
	tests := []struct {
		name    string
		status  MissingStatus
		wantErr bool
		errMsg  string
	}{
		{
			name:    "from found to reunited",
			status:  MissingStatusFound,
			wantErr: false,
		},
		{
			name:    "from missing cannot reunite",
			status:  MissingStatusMissing,
			wantErr: true,
			errMsg:  "只有已找到状态才能标记为团聚",
		},
		{
			name:    "from reunited cannot reunite again",
			status:  MissingStatusReunited,
			wantErr: true,
			errMsg:  "只有已找到状态才能标记为团聚",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MissingPerson{Status: tt.status}
			err := mp.MarkReunited()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, MissingStatusReunited, mp.Status)
		})
	}
}

func TestMissingPerson_Close(t *testing.T) {
	mp := &MissingPerson{Status: MissingStatusSearching}
	
	err := mp.Close("案件已解决")
	require.NoError(t, err)
	assert.Equal(t, MissingStatusClosed, mp.Status)
	
	// Cannot close again
	err = mp.Close("再次关闭")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "案件已关闭")
}

func TestMissingPerson_AssignTo(t *testing.T) {
	userID := "user-id"
	mp := &MissingPerson{}
	
	mp.AssignTo(userID)
	assert.Equal(t, &userID, mp.AssignedTo)
}

func TestMissingPerson_IncrementViews(t *testing.T) {
	mp := &MissingPerson{Views: 0}
	
	mp.IncrementViews()
	assert.Equal(t, 1, mp.Views)
	
	mp.IncrementViews()
	assert.Equal(t, 2, mp.Views)
}

func TestMissingPerson_GetAgeAtMissing(t *testing.T) {
	// Use fixed dates for predictable age calculation
	missingTime := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	
	mp := &MissingPerson{
		BirthDate:   &birthDate,
		MissingTime: missingTime,
	}
	
	age := mp.GetAgeAtMissing()
	assert.Equal(t, 30, age) // 2020 - 1990 = 30 (June > January, so age is 30)
	
	// Without birth date, use Age field
	mp.BirthDate = nil
	mp.Age = 25
	age = mp.GetAgeAtMissing()
	assert.Equal(t, 25, age)
}
