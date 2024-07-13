package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/whatsauth/itmodel"
)

func GetHome(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	resp.Response = helper.GetIPaddress()
	helper.WriteResponse(respw, http.StatusOK, resp)
}

func PostInboxNomor(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	var msg itmodel.IteungMessage
	waphonenumber := helper.GetParam(req)
	prof, err := helper.GetAppProfile(waphonenumber, config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		helper.WriteResponse(respw, http.StatusServiceUnavailable, resp)
		return
	}
	if helper.GetSecretFromHeader(req) == prof.Secret {
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			resp.Response = err.Error()
			helper.WriteResponse(respw, http.StatusBadRequest, resp)
			return
		} else if msg.Message != "" {
			_, err = helper.InsertOneDoc(config.Mongoconn, "inbox", msg)
			if err != nil {
				resp.Response = err.Error()
			}
			resp, err = helper.WebHook(prof.QRKeyword, waphonenumber, config.WAAPIQRLogin, config.WAAPIMessageText, msg, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
			}
			helper.WriteResponse(respw, http.StatusOK, resp)
			return
		} else {
			resp.Response = "pesan kosong"
			helper.WriteResponse(respw, http.StatusOK, resp)
			return
		}
	}
	resp.Response = "Wrong Secret"
	helper.WriteResponse(respw, http.StatusForbidden, resp)
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	httpstatus := http.StatusServiceUnavailable
	//prof, err := helper.GetAppProfile(config.WAPhoneNumber, config.Mongoconn)
	profs, err := helper.GetAllDoc[[]itmodel.Profile](config.Mongoconn, "profile")
	if err != nil {
		resp.Response = err.Error()
	} else {
		for _, prof := range profs {
			dt := &itmodel.WebHook{
				URL:    prof.URL,
				Secret: prof.Secret,
			}
			res, err := helper.RefreshToken(dt, prof.Phonenumber, config.WAAPIGetToken, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
				break
			} else {
				resp.Response = helper.Jsonstr(res.ModifiedCount)
				httpstatus = http.StatusOK
			}
		}
	}

	helper.WriteResponse(respw, httpstatus, resp)
}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	resp.Response = "Not Found"
	helper.WriteResponse(respw, http.StatusNotFound, resp)
}
