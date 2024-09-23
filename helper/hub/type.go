package hub

import "time"

type SessionHub struct {
	ID         string    `bson:"_id,omitempty"`
	UserPhone  string    `bson:"userphone"`
	UserName   string    `bson:"username"`
	AdminPhone string    `bson:"adminphone"`
	AdminName  string    `bson:"adminname"`
	CreatedAt  time.Time `bson:"createdAt"`
}
