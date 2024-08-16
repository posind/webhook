package passwordhash

import (
	"encoding/json"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/watoken"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func EncodeToken(email, privatekey string) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(2 * time.Hour))
	token.SetString("user", email)
	key, err := paseto.NewV4AsymmetricSecretKeyFromHex(privatekey)
	return token.V4Sign(key, nil), err
}

func Decoder(publickey, tokenstr string) (payload Payload, err error) {
	var token *paseto.Token
	var pubKey paseto.V4AsymmetricPublicKey

	pubKey, err = paseto.NewV4AsymmetricPublicKeyFromHex(publickey)
	if err != nil {
		return payload, fmt.Errorf("failed to create public key: %s", err)
	}

	parser := paseto.NewParser()
	token, err = parser.ParseV4Public(pubKey, tokenstr, nil)
	if err != nil {
		return payload, fmt.Errorf("failed to parse token: %s", err)
	}

	// Print the raw claims JSON
	fmt.Printf("Token claims JSON: %s\n", string(token.ClaimsJSON()))

	err = json.Unmarshal(token.ClaimsJSON(), &payload)
	if err != nil {
		return payload, fmt.Errorf("failed to unmarshal token claims: %s", err)
	}

	return payload, nil
}

func DecodeGetUser(PublicKey, tokenStr string) (pay string, err error) {
	key, err := Decoder(PublicKey, tokenStr)
	if err != nil {
		fmt.Println("Cannot decode the token", err.Error())
		return "", err // Mengembalikan nilai kosong dan informasi kesalahan
	}

	// Use the extracted ID to fetch the username from the database
	return key.ID, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func TokenEncoder(username, privatekey string) string {
	resp := new(ResponseEncode)
	encode, err := watoken.Encode(username, privatekey)
	if err != nil {
		resp.Message = "Gagal Encode" + err.Error()
	} else {
		resp.Token = encode
		resp.Message = "Welcome"
	}

	return at.Jsonstr(resp)
}

func PasswordValidator(loginReq model.LoginRequest) bool {
	// Mencari user berdasarkan email
	filter := bson.M{"email": loginReq.Email}
	data, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", filter)

	// Jika terjadi error atau user tidak ditemukan, kembalikan false
	if err != nil || data.Email == "" {
		return false
	}

	// Memeriksa apakah password yang diberikan sesuai dengan hash yang tersimpan
	hashChecker := CheckPasswordHash(loginReq.Password, data.Password)
	return hashChecker
}
