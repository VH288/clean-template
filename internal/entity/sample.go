package entity

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

func (s Sample) Validate() error {
	return validator.New().Struct(s)
}
