package models

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusBanned    UserStatus = "banned"
	UserStatusDeleted   UserStatus = "deleted"
	UserStatusSuspended UserStatus = "suspended"
)

// User represents a core user account
type User struct {
	UserID    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"user_id"`
	Status    UserStatus `gorm:"size:32;not null;default:'active';index:idx_users_status" json:"status"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_users_created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Profile *Profile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"profile,omitempty"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsBanned checks if the user account is banned
func (u *User) IsBanned() bool {
	return u.Status == UserStatusBanned
}

// Profile represents user profile information
type Profile struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Username    *string   `gorm:"size:30;uniqueIndex:uq_profiles_username" json:"username,omitempty"`
	DisplayName *string   `gorm:"size:50" json:"display_name,omitempty"`
	AvatarURL   *string   `gorm:"type:text" json:"avatar_url,omitempty"`
	BannerURL   *string   `gorm:"type:text" json:"banner_url,omitempty"`
	Bio         *string   `gorm:"type:text" json:"bio,omitempty"`
	Locale      string    `gorm:"size:10;default:'en'" json:"locale"`
	Timezone    string    `gorm:"size:50;default:'UTC'" json:"timezone"`
	SocialsJSON *string   `gorm:"type:jsonb;column:socials_json" json:"socials_json,omitempty"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_profiles_updated_at" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Profile) TableName() string {
	return "profiles"
}

// HasUsername checks if the profile has a username set
func (p *Profile) HasUsername() bool {
	return p.Username != nil && *p.Username != ""
}
