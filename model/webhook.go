package model

import (
	"time"

	"github.com/gocroot/helper/gcallapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Laporan struct {
	ID          primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty" query:"id" url:"_id,omitempty" reqHeader:"_id"`
	MeetID      primitive.ObjectID   `json:"meetid,omitempty" bson:"meetid,omitempty"`
	MeetEvent   gcallapi.SimpleEvent `json:"meetevent,omitempty" bson:"meetevent,omitempty"`
	Project     Project              `json:"project,omitempty" bson:"project,omitempty"`
	User        Userdomyikado        `json:"user,omitempty" bson:"user,omitempty"`
	Petugas     string               `json:"petugas,omitempty" bson:"petugas,omitempty"`
	NoPetugas   string               `json:"nopetugas,omitempty" bson:"nopetugas,omitempty"`
	Kode        string               `json:"kode,omitempty" bson:"kode,omitempty"`
	Team        string               `json:"team,omitempty" bson:"team,omitempty"`
	Scope       string               `json:"scope,omitempty" bson:"scope,omitempty"`
	Nama        string               `json:"nama,omitempty" bson:"nama,omitempty"`
	Phone       string               `json:"phone,omitempty" bson:"phone,omitempty"`
	Masalah     string               `json:"masalah,omitempty" bson:"masalah,omitempty"`
	Solusi      string               `json:"solusi,omitempty" bson:"solusi,omitempty"`
	Komentar    string               `json:"komentar,omitempty" bson:"komentar,omitempty"`
	Terlayani   bool                 `json:"terlayani,omitempty" bson:"terlayani,omitempty"`
	Rating      float64              `json:"rating,omitempty" bson:"rating,omitempty"`
	RateLayanan int                  `json:"ratelayanan,omitempty" bson:"ratelayanan,omitempty"`
}

type PushReport struct {
	ProjectName string        `bson:"projectname" json:"projectname"`
	Project     Project       `bson:"project" json:"project"`
	User        Userdomyikado `bson:"user,omitempty" json:"user,omitempty"`
	Username    string        `bson:"username" json:"username"`
	Email       string        `bson:"email,omitempty" json:"email,omitempty"`
	Repo        string        `bson:"repo" json:"repo"`
	Ref         string        `bson:"ref" json:"ref"`
	Message     string        `bson:"message" json:"message"`
	Modified    string        `bson:"modified,omitempty" json:"modified,omitempty"`
	RemoteAddr  string        `bson:"remoteaddr,omitempty" json:"remoteaddr,omitempty"`
}

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Secret      string             `bson:"secret" json:"secret"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Owner       Userdomyikado      `bson:"owner" json:"owner"`
	WAGroupID   string             `bson:"wagroupid,omitempty" json:"wagroupid,omitempty"`
	RepoOrg     string             `bson:"repoorg,omitempty" json:"repoorg,omitempty"`
	RepoLogName string             `bson:"repologname,omitempty" json:"repologname,omitempty"`
	Members     []Userdomyikado    `bson:"members,omitempty" json:"members,omitempty"`
	Closed      bool               `bson:"closed,omitempty" json:"closed,omitempty"`
}

type Userdomyikado struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name                 string             `bson:"name,omitempty" json:"name,omitempty"`
	PhoneNumber          string             `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	Email                string             `bson:"email,omitempty" json:"email,omitempty"`
	GithubUsername       string             `bson:"githubusername,omitempty" json:"githubusername,omitempty"`
	GitlabUsername       string             `bson:"gitlabusername,omitempty" json:"gitlabusername,omitempty"`
	GitHostUsername      string             `bson:"githostusername,omitempty" json:"githostusername,omitempty"`
	Poin                 float64            `bson:"poin,omitempty" json:"poin,omitempty"`
	GoogleProfilePicture string             `bson:"googleprofilepicture,omitempty" json:"picture,omitempty"`
	Team                 string             `json:"team,omitempty" bson:"team,omitempty"`
	Scope                string             `json:"scope,omitempty" bson:"scope,omitempty"`
	Section              string             `json:"section,omitempty" bson:"section,omitempty"`
	Chapter              string             `json:"chapter,omitempty" bson:"chapter,omitempty"`
	LinkedDevice         string             `json:"linkeddevice,omitempty" bson:"linkeddevice,omitempty"`
	JumlahAntrian        int                `json:"jumlahantrian,omitempty" bson:"jumlahantrian,omitempty"`
}

type Task struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ProjectID string             `bson:"projectid" json:"projectid"`
	Name      string             `bson:"name" json:"name"`
	PIC       Userdomyikado      `bson:"pic" json:"pic"`
	Done      bool               `bson:"done,omitempty" json:"done,omitempty"`
}

type Stp struct {
	PhoneNumber  string    `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	PasswordHash string    `bson:"password,omitempty" json:"password,omitempty"`
	CreatedAt    time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
}

type VerifyRequest struct {
	PhoneNumber string `json:"phonenumber"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phonenumber"`
	Password    string `json:"password"`
}