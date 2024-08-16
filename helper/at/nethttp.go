package domyApi

import (
	"encoding/json"
	"log"
	"net/http"
)

func GetSecretFromHeader(r *http.Request) (secret string) {
	if r.Header.Get("secret") != "" {
		secret = r.Header.Get("secret")
	} else if r.Header.Get("Secret") != "" {
		secret = r.Header.Get("Secret")
	}
	return
}

func GetLoginFromHeader(r *http.Request) (secret string) {
	if r.Header.Get("login") != "" {
		secret = r.Header.Get("login")
	} else if r.Header.Get("Login") != "" {
		secret = r.Header.Get("Login")
	}
	return
}

func Jsonstr(strc interface{}) string {
	jsonData, err := json.Marshal(strc)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonData)
}

func WriteJSON(respw http.ResponseWriter, statusCode int, content interface{}) {
	respw.Header().Set("Content-Type", "application/json")
	respw.WriteHeader(statusCode)
	respw.Write([]byte(Jsonstr(content)))
}

func WriteString(respw http.ResponseWriter, statusCode int, content string) {
	respw.WriteHeader(statusCode)
	respw.Write([]byte(content))
}
