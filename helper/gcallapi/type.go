package gcallapi

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CredentialRecord struct {
	Token               string    `bson:"token"`
	RefreshToken        string    `bson:"refresh_token"`
	TokenURI            string    `bson:"token_uri"`
	ClientID            string    `bson:"client_id"`
	ClientSecret        string    `bson:"client_secret"`
	Expiry              time.Time `bson:"expiry"`
	AuthProviderCertURL string    `bson:"auth_provider_x509_cert_url"`
	AuthURI             string    `bson:"auth_uri"`
	ProjectID           string    `bson:"project_id"`
	RedirectURIs        []string  `bson:"redirect_uris"`
	JavascriptOrigins   []string  `bson:"javascript_origins"`
	Scopes              []string  `bson:"scopes"`
}

type Attachment struct {
	FileUrl  string `json:"fileurl,omitempty" bson:"fileurl,omitempty"`
	MimeType string `json:"mimetype,omitempty" bson:"mimetype,omitempty"`
	Title    string `json:"title,omitempty" bson:"title,omitempty"`
}

type SimpleEvent struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ProjectID   primitive.ObjectID `json:"project_id,omitempty" bson:"project_id,omitempty"`
	Summary     string             `json:"summary,omitempty" bson:"summary,omitempty"`
	Location    string             `json:"location,omitempty" bson:"location,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Date        string             `json:"date,omitempty" bson:"date,omitempty"`           // YYYY-MM-DD
	TimeStart   string             `json:"timestart,omitempty" bson:"timestart,omitempty"` // HH:MM:SS
	TimeEnd     string             `json:"timeend,omitempty" bson:"timeend,omitempty"`     // HH:MM:SS
	Attendees   []string           `json:"attendees,omitempty" bson:"attendees,omitempty"`
	Attachments []Attachment       `json:"attachments,omitempty" bson:"attachments,omitempty"` // New field for attachments

}
