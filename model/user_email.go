package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" `
	Username string             `json:"username" bson:"username"`
	Email    string             `bson:"email,omitempty" json:"email,omitempty"`
	Password string             `json:"password" bson:"password"`
	Token    string             `json:"token,omitempty" bson:"token,omitempty"`
	Private  string             `json:"private,omitempty" bson:"private,omitempty"`
	Public   string             `json:"public,omitempty" bson:"public,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Credential struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}
