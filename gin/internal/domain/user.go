package domain

import (
	"time"
)

type User struct {
	// Mapped explicitly to database column names, keeping clean JSON field names for Gin
	ID           int32     `gorm:"primaryKey;column:LoginId" json:"id"`
	Username     string    `gorm:"column:Username" json:"username"`
	Password     string    `gorm:"column:Password" json:"-"` // "json:"-"" hides the hashed password from API JSON outputs
	GroupId      int32     `gorm:"column:GroupId" json:"group_id"`
	Email        string    `gorm:"column:email" json:"email"`
	PhoneNumber  string    `gorm:"column:PhoneNumber" json:"phone_number"`
	ImgUrl       string    `gorm:"column:ImgUrl" json:"img_url"`
	FlagUse      bool      `gorm:"column:FlagUse" json:"flag_use"`
	DateStart    time.Time `gorm:"column:DateStart" json:"date_start"`
	DateEnd      time.Time `gorm:"column:DateEnd" json:"date_end"`
	LoginDesc    string    `gorm:"column:LoginDesc" json:"login_desc"`
	LastLogin    time.Time `gorm:"column:LastLogin" json:"last_login"`
	DepartmentId int32     `gorm:"column:DepartmentId" json:"department_id"`
	IsWarehouse  bool      `gorm:"column:IsWarehouse" json:"is_warehouse"`
}
