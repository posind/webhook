package tiket

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bantuan struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserName     string             `bson:"username,omitempty" json:"username,omitempty"`
	UserPhone    string             `bson:"userphone,omitempty" json:"userphone,omitempty"`
	AdminPhone   string             `bson:"adminphone,omitempty" json:"adminphone,omitempty"`
	AdminName    string             `bson:"adminname,omitempty" json:"adminname,omitempty"`
	Prov         string             `bson:"prov,omitempty" json:"prov,omitempty"`
	KabKot       string             `bson:"kabkot,omitempty" json:"kabkot,omitempty"`
	Kec          string             `bson:"kec,omitempty" json:"kec,omitempty"`
	Desa         string             `bson:"desa,omitempty" json:"desa,omitempty"`
	UserMessage  string             `bson:"usermessage,omitempty" json:"usermessage,omitempty"`
	AdminMessage string             `bson:"adminmessage,omitempty" json:"adminmessage,omitempty"`
	StartAt      time.Time          `bson:"startat,omitempty" json:"startat,omitempty"`
	ResponsAt    time.Time          `bson:"responsat,omitempty" json:"responsat,omitempty"`
	CloseAt      time.Time          `bson:"closeat,omitempty" json:"closeat,omitempty"`
	Terlayani    bool               `json:"terlayani,omitempty" bson:"terlayani,omitempty"`
	RateLayanan  int                `json:"ratelayanan,omitempty" bson:"ratelayanan,omitempty"`
	Masukan      string             `json:"masukan,omitempty" bson:"masukan,omitempty"`
}
