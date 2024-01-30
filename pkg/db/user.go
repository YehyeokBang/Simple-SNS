package db

import "time"

type User struct {
	ID        uint
	UserId    string `gorm:"type:varchar(100);unique"`
	Password  string `gorm:"type:varchar(100)"`
	Name      string `gorm:"type:varchar(100)"`
	Age       uint32
	Sex       string `gorm:"type:varchar(100)"`
	Birthday  *time.Time
	Introduce string `gorm:"type:varchar(100)"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Posts []Post
}
