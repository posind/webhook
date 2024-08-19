package helpdesk

import "go.mongodb.org/mongo-driver/bson/primitive"

type Helpdesk struct {
	Team          string `json:"team,omitempty" bson:"team,omitempty"`
	Scope         string `json:"scope,omitempty" bson:"scope,omitempty"`
	Section       string `json:"section,omitempty" bson:"section,omitempty"`
	Chapter       string `json:"chapter,omitempty" bson:"chapter,omitempty"`
	Name          string `json:"name,omitempty" bson:"name,omitempty"`
	Phonenumbers  string `json:"phonenumbers,omitempty" bson:"phonenumbers,omitempty"`
	JumlahAntrian int    `json:"jumlahantrian,omitempty" bson:"jumlahantrian,omitempty"`
}

type User struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Team         string             `json:"team,omitempty" bson:"team,omitempty"`
	Scope        string             `json:"scope,omitempty" bson:"scope,omitempty"`
	Name         string             `json:"name,omitempty" bson:"name,omitempty"`
	Phonenumbers string             `json:"phonenumbers,omitempty" bson:"phonenumbers,omitempty"`
	Terlayani    bool               `json:"terlayani,omitempty" bson:"terlayani,omitempty"`
	Masalah      string             `json:"masalah,omitempty" bson:"masalah,omitempty"`
	Solusi       string             `json:"solusi,omitempty" bson:"solusi,omitempty"`
	RateLayanan  int                `json:"ratelayanan,omitempty" bson:"ratelayanan,omitempty"`
	Operator     Helpdesk           `json:"operator,omitempty" bson:"operator,omitempty"`
}
