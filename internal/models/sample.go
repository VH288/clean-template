package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Sample struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(100)" validate:"required"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (*Sample) TableName() string {
	return "sample"
}

func (l Sample) Validate() error {
	v := validator.New()
	return v.Struct(l)
}

type SampleParam struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
