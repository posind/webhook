package menu

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID          string     `bson:"_id,omitempty"`
	PhoneNumber string     `bson:"phonenumber"`
	Menulist    []MenuList `bson:"list"`
	CreatedAt   time.Time  `bson:"createdAt"`
}

type Menu struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"` // Field ID untuk MongoDB
	Keyword string             `bson:"keyword,omitempty" json:"keyword,omitempty"`
	Header  string             `bson:"header,omitempty" json:"header,omitempty"`
	List    []MenuList         `bson:"list,omitempty" json:"list,omitempty"`
	Footer  string             `bson:"footer,omitempty" json:"footer,omitempty"`
}

type MenuList struct {
	No      int    `bson:"no,omitempty" json:"no,omitempty"`
	Keyword string `bson:"keyword,omitempty" json:"keyword,omitempty"`
	Konten  string `bson:"konten,omitempty" json:"konten,omitempty"`
}
