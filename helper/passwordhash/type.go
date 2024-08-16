package passwordhash

import "time"

type ResponseEncode struct {
	Message string `json:"message,omitempty" bson:"message,omitempty"`
	Token   string `json:"token,omitempty" bson:"token,omitempty"`
}

type Payload struct {
	User string    `json:"user"`
	Exp  time.Time `json:"exp"`
	Iat  time.Time `json:"iat"`
	Nbf  time.Time `json:"nbf"`
}
