package db

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID              uint `gorm:"primaryKey"`
	UserID          uint
	User            User
	PostID          uint
	Post            Post
	Content         string `gorm:"type:varchar(500)"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt
	ParentCommentID *uint     // 부모 댓글의 ID를 저장
	ParentComment   *Comment  // 부모 댓글을 참조
	ChildComments   []Comment `gorm:"foreignkey:ParentCommentID"`
}
