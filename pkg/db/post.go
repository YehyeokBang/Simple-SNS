package db

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	User      User
	Content   string `gorm:"type:varchar(500)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  gorm.DeletedAt

	Comments []Comment `gorm:"foreignKey:PostID"`
}
