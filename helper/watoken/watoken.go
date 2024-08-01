package watoken

import (
	"encoding/json"
	"time"

	"aidanwoods.dev/go-paseto"
)

type Payload[T any] struct {
	Id    string    `json:"id"`
	Alias string    `json:"alias"`
	Exp   time.Time `json:"exp"`
	Iat   time.Time `json:"iat"`
	Nbf   time.Time `json:"nbf"`
	Data  T         `json:"data"`
}

func GenerateKey() (privateKey, publicKey string) {
	secretKey := paseto.NewV4AsymmetricSecretKey() // don't share this!!!
	publicKey = secretKey.Public().ExportHex()     // DO share this one
	privateKey = secretKey.ExportHex()
	return privateKey, publicKey
}

func Encode(id string, privateKey string) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(2 * time.Hour))
	token.SetString("id", id)
	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	return token.V4Sign(secretKey, nil), err

}

func EncodeWithStruct[T any](id string, data *T, privateKey string) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(2 * time.Hour))
	token.SetString("id", id)

	err := token.Set("data", data)
	if err != nil {
		return "", err
	}

	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	return token.V4Sign(secretKey, nil), err

}

func EncodeWithStructDuration[T any](id string, data *T, privateKey string, dur ...time.Duration) (string, error) {
	duration := time.Duration(2 * time.Hour)
	if len(dur) > 0 {
		duration = dur[0]
	}

	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(duration))
	token.SetString("id", id)

	err := token.Set("data", data)
	if err != nil {
		return "", err
	}

	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	return token.V4Sign(secretKey, nil), err

}

func EncodeforHours(id, alias, privateKey string, hours int32) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(time.Duration(hours) * time.Hour))
	token.SetString("id", id)
	token.SetString("alias", alias)
	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	return token.V4Sign(secretKey, nil), err

}

func EncodeforMinutes(id string, privateKey string, minutes int32) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(time.Duration(minutes) * time.Minute))
	token.SetString("id", id)
	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	return token.V4Sign(secretKey, nil), err

}

func EncodeforSeconds(id string, privateKey string, seconds int32) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(time.Duration(seconds) * time.Second))
	token.SetString("id", id)
	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	return token.V4Sign(secretKey, nil), err

}

func Decode(publicKey string, tokenstring string) (payload Payload[any], err error) {
	var token *paseto.Token
	var pubKey paseto.V4AsymmetricPublicKey
	pubKey, err = paseto.NewV4AsymmetricPublicKeyFromHex(publicKey) // this wil fail if given key in an invalid format
	if err != nil {
		return
	}

	parser := paseto.NewParser()                                // only used because this example token has expired, use NewParser() (which checks expiry by default)
	token, err = parser.ParseV4Public(pubKey, tokenstring, nil) // this will fail if parsing failes, cryptographic checks fail, or validation rules fail
	if err != nil {
		return
	}

	err = json.Unmarshal(token.ClaimsJSON(), &payload)
	return
}
func DecodeWithStruct[T any](publicKey string, tokenstring string) (payload Payload[T], err error) {
	pubKey, err := paseto.NewV4AsymmetricPublicKeyFromHex(publicKey) // this wil fail if given key in an invalid format
	if err != nil {
		return
	}

	parser := paseto.NewParser()                                 // only used because this example token has expired, use NewParser() (which checks expiry by default)
	token, err := parser.ParseV4Public(pubKey, tokenstring, nil) // this will fail if parsing failes, cryptographic checks fail, or validation rules fail
	if err != nil {
		return
	}

	err = json.Unmarshal(token.ClaimsJSON(), &payload)
	return
}

func DecodeGetId(publicKey string, tokenstring string) (string, error) {
	payload, err := Decode(publicKey, tokenstring)
	if err != nil {
		return "", err
	}
	return payload.Id, nil
}