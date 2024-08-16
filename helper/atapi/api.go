package atapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func PostStructWithToken[T any](tokenkey string, tokenvalue string, structname interface{}, urltarget string) (result T, err error) {
	client := http.Client{}
	mJson, _ := json.Marshal(structname)
	req, err := http.NewRequest("POST", urltarget, bytes.NewBuffer(mJson))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(tokenkey, tokenvalue)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		rawstring := string(respBody)
		err = errors.New("Not A Valid JSON Response from " + urltarget + ". CONTENT: " + rawstring)
		return
	}
	return
}

func PostStructWithTokenMod[T any](tokenkey string, tokenvalue string, structname interface{}, urltarget string) (statusCode int, result T, err error) {
	client := http.Client{}
	mJson, _ := json.Marshal(structname)
	req, err := http.NewRequest("POST", urltarget, bytes.NewBuffer(mJson))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(tokenkey, tokenvalue)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	statusCode = resp.StatusCode
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		rawstring := string(respBody)
		err = errors.New("Not A Valid JSON Response from " + urltarget + ". CONTENT: " + rawstring)
		return
	}
	return
}

func Get[T any](urltarget string) (statusCode int, result T, err error) {
	resp, err := http.Get(urltarget)
	if err != nil {
		return
	}
	statusCode = resp.StatusCode
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if er := json.Unmarshal(body, &result); er != nil {
		return
	}
	return
}
