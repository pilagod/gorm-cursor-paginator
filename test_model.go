package paginator

import "time"

type order struct {
	ID        int       `gorm:"primary_key"`
	Name      *string   `gorm:"type:varchar(30)"`
	Items     []item    `gorm:"foreignkey:OrderID"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}

type item struct {
	ID      int   `gorm:"primary_key"`
	OrderID int   `gorm:"type:integer;not null"`
	Order   Order `gorm:"foreignkey:OrderID"`
}
