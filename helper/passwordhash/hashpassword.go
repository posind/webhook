package passwordhash

import (
	"encoding/json"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gocroot/helper/at"
	"github.com/whatsauth/watoken"
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

func TokenEncoder(phoneNumber, privatekey string) string {
	resp := new(ResponseEncode)
	encode, err := watoken.Encode(phoneNumber, privatekey)
	if err != nil {
		resp.Message = "Gagal Encode: " + err.Error()
	} else {
		resp.Token = encode
		resp.Message = "Welcome"
	}

	return at.Jsonstr(resp)
}
