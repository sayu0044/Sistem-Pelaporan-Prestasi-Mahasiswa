package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUser_BeforeCreate(t *testing.T) {
	tests := []struct {
		name     string
		user     *model.User
		wantNil  bool
		checkID  bool
	}{
		{
			name: "ID is nil should generate UUID",
			user: &model.User{
				ID: uuid.Nil,
			},
			wantNil: true,
			checkID: true,
		},
		{
			name: "ID already exists should not change",
			user: &model.User{
				ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantNil: true,
			checkID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.user.ID
			var tx *gorm.DB

			err := tt.user.BeforeCreate(tx)

			assert.NoError(t, err)
			if tt.checkID {
				assert.NotEqual(t, uuid.Nil, tt.user.ID)
				assert.NotEqual(t, originalID, tt.user.ID)
			} else {
				assert.Equal(t, originalID, tt.user.ID)
			}
		})
	}
}

