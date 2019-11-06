package orm

import (
	"github.com/satori/go.uuid"
	"time"

	"github.com/jinzhu/gorm"
)

type GormModel struct {
	ID        string     `database:"primary_key;type:varchar(255);" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index"`
}

func (model *GormModel) BeforeCreate(scope *gorm.Scope) error {
	u1 := uuid.Must(uuid.NewV4(), nil)
	scope.SetColumn("ID", u1.String())
	return nil
}
