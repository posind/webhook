package lms

import (
	"encoding/json"
	"time"
)

type CustomTime time.Time

const customTimeFormat = "2006-01-02T15:04:05Z07:00"

// UnmarshalJSON handles both RFC3339 and Unix timestamp formats
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	str := string(b)

	if str == "null" || str == "" {
		return nil
	}

	// Remove surrounding quotes for string values
	if str[0] == '"' {
		str = str[1 : len(str)-1]
	}

	// Try parsing as RFC3339 format
	t, err := time.Parse(customTimeFormat, str)
	if err == nil {
		*ct = CustomTime(t)
		return nil
	}

	// Try parsing as Unix timestamp
	var ts int64
	if err := json.Unmarshal([]byte(str), &ts); err == nil {
		*ct = CustomTime(time.Unix(ts, 0).UTC())
		return nil
	}

	return err
}

// MarshalJSON formats the time in RFC3339 format
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(ct).Format(customTimeFormat))
}

type UnixTime struct {
	time.Time
}

// UnmarshalJSON parses a Unix timestamp
func (ut *UnixTime) UnmarshalJSON(b []byte) error {
	var ts int64
	if err := json.Unmarshal(b, &ts); err != nil {
		return err
	}
	ut.Time = time.Unix(ts, 0).UTC()
	return nil
}

// MarshalJSON converts the time to Unix timestamp
func (ut UnixTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ut.Unix())
}

type LoginProfile struct {
	Username  string `bson:"user,omitempty"`
	Bearer    string `bson:"bearer,omitempty"`
	Xsrf      string `bson:"xsrf,omitempty"`
	Lsession  string `bson:"lsession,omitempty"`
	URLXlsx   string `bson:"urlxlsx,omitempty"`
	URLUsers  string `bson:"urlusers,omitempty"`
	URLCookie string `bson:"urlcookie,omitempty"`
}

type Position struct {
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	ParentID  string      `json:"parent_id,omitempty"`
	Order     *int        `json:"order,omitempty"`
	IsDelete  bool        `json:"is_delete,omitempty"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
}

type Province struct {
	Kode      string      `json:"kode,omitempty"`
	Nama      string      `json:"nama,omitempty"`
	IsDelete  bool        `json:"is_delete,omitempty"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
	IDs       string      `json:"ids,omitempty"`
}

type Regency struct {
	Kode      string      `json:"kode,omitempty"`
	Nama      string      `json:"nama,omitempty"`
	IsDelete  bool        `json:"is_delete,omitempty"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
	IDs       *string     `json:"ids,omitempty"`
}

type District struct {
	Kode      string      `json:"kode,omitempty"`
	Nama      string      `json:"nama,omitempty"`
	IsDelete  bool        `json:"is_delete,omitempty"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
	IDs       *string     `json:"ids,omitempty"`
}

type Village struct {
	Kode      string      `json:"kode,omitempty"`
	Nama      string      `json:"nama,omitempty"`
	IsDelete  bool        `json:"is_delete,omitempty"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
	IDs       *string     `json:"ids,omitempty"`
}

type UserProfile struct {
	TMT         string   `json:"tmt,omitempty"`
	Position    Position `json:"position,omitempty"`
	Province    Province `json:"province,omitempty"`
	Regency     Regency  `json:"regency,omitempty"`
	District    District `json:"district,omitempty"`
	Village     Village  `json:"village,omitempty"`
	Decree      string   `json:"decree,omitempty"`
	TrainerCert *string  `json:"trainer_cert,omitempty"`
}

type ApprovedBy struct {
	ID               string      `json:"id,omitempty"`
	Fullname         string      `json:"fullname,omitempty"`
	Username         string      `json:"username,omitempty"`
	Phone            string      `json:"phone,omitempty"`
	Email            string      `json:"email,omitempty"`
	EmailVerifiedAt  *UnixTime   `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt  *UnixTime   `json:"phone_verified_at,omitempty"`
	VerificationCode *string     `json:"verification_code,omitempty"`
	ProfileVerified  bool        `json:"profile_verified,omitempty"`
	ProfileApproved  int         `json:"profile_approved,omitempty"`
	LastLoginAt      *UnixTime   `json:"last_login_at,omitempty"`
	IsDelete         bool        `json:"is_delete,omitempty"`
	CreatedBy        string      `json:"created_by,omitempty"`
	UpdatedBy        *string     `json:"updated_by,omitempty"`
	CreatedAt        *CustomTime `json:"created_at,omitempty"`
	UpdatedAt        *CustomTime `json:"updated_at,omitempty"`
	FcmToken         *string     `json:"fcm_token,omitempty"`
	DeletedAt        *CustomTime `json:"deleted_at,omitempty"`
	ApprovedBy       *string     `json:"approved_by,omitempty"`
	RejectedBy       *string     `json:"rejected_by,omitempty"`
	ApprovedAt       *CustomTime `json:"approved_at,omitempty"`
	RejectedAt       *CustomTime `json:"rejected_at,omitempty"`
}

type RejectedBy struct {
	ID               string      `json:"id,omitempty"`
	Fullname         string      `json:"fullname,omitempty"`
	Username         string      `json:"username,omitempty"`
	Phone            string      `json:"phone,omitempty"`
	Email            string      `json:"email,omitempty"`
	EmailVerifiedAt  *UnixTime   `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt  *UnixTime   `json:"phone_verified_at,omitempty"`
	VerificationCode *string     `json:"verification_code,omitempty"`
	ProfileVerified  bool        `json:"profile_verified,omitempty"`
	ProfileApproved  int         `json:"profile_approved,omitempty"`
	LastLoginAt      *UnixTime   `json:"last_login_at,omitempty"`
	IsDelete         bool        `json:"is_delete,omitempty"`
	CreatedBy        string      `json:"created_by,omitempty"`
	UpdatedBy        *string     `json:"updated_by,omitempty"`
	CreatedAt        *CustomTime `json:"created_at,omitempty"`
	UpdatedAt        *CustomTime `json:"updated_at,omitempty"`
	FcmToken         *string     `json:"fcm_token,omitempty"`
	DeletedAt        *CustomTime `json:"deleted_at,omitempty"`
	ApprovedBy       *string     `json:"approved_by,omitempty"`
	RejectedBy       *string     `json:"rejected_by,omitempty"`
	ApprovedAt       *CustomTime `json:"approved_at,omitempty"`
	RejectedAt       *CustomTime `json:"rejected_at,omitempty"`
}

type User struct {
	ID              string       `json:"id,omitempty"`
	Fullname        string       `json:"fullname,omitempty"`
	Username        string       `json:"username,omitempty"`
	Email           string       `json:"email,omitempty"`
	EmailVerified   *UnixTime    `json:"email_verified,omitempty"`
	ProfileVerified bool         `json:"profile_verified,omitempty"`
	ProfileApproved int          `json:"profile_approved,omitempty"`
	LastLoginAt     *UnixTime    `json:"last_login_at,omitempty"`
	UserProfile     *UserProfile `json:"user_profile,omitempty"`
	CreatedAt       *CustomTime  `json:"created_at,omitempty"`
	Roles           []string     `json:"roles,omitempty"`
	ApprovedBy      *ApprovedBy  `json:"approved_by,omitempty"`
	ApprovedAt      *CustomTime  `json:"approved_at,omitempty"`
	RejectedBy      *RejectedBy  `json:"rejected_by,omitempty"`
	RejectedAt      *CustomTime  `json:"rejected_at,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page,omitempty"`
	FirstItem   int `json:"first_item,omitempty"`
	LastItem    int `json:"last_item,omitempty"`
	LastPage    int `json:"last_page,omitempty"`
	Total       int `json:"total,omitempty"`
}

type Data struct {
	Data []User `json:"data,omitempty"`
	Meta Meta   `json:"meta,omitempty"`
}

type Root struct {
	Data Data `json:"data,omitempty"`
}

// 1. Belum Lengkap
// 2. Menunggu Persetujuan
// 3. Disetujui
// 4. Ditolak
type RekapitulasiUser struct {
	BelumLengkap        int64 `json:"belumlengkap,omitempty"`
	MenungguPersetujuan int64 `json:"menunggupersetujuan,omitempty"`
	Disetujui           int64 `json:"disetujui,omitempty"`
	Ditolak             int64 `json:"ditolak,omitempty"`
	Total               int64 `json:"total,omitempty"`
}

type ResponseAPIPD struct {
	Success bool   `json:"success"`
	Data    DataPD `json:"data"`
}

type ContactAdmin struct {
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
}

type DataPD struct {
	Fullname             string         `json:"fullname"`
	Province             string         `json:"province"`
	Regency              string         `json:"regency"`
	District             string         `json:"district"`
	Village              string         `json:"village"`
	ContactAdminRegency  []ContactAdmin `json:"contact_admin_regency"`
	ContactAdminProvince []ContactAdmin `json:"contact_admin_province"`
}
