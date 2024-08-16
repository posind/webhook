package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetHome(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	resp.Response = at.GetIPaddress()
	at.WriteResponse(respw, http.StatusOK, resp)
}

func PostInboxNomor(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	var msg itmodel.IteungMessage
	waphonenumber := at.GetParam(req)
	prof, err := at.GetAppProfile(waphonenumber, config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteResponse(respw, http.StatusServiceUnavailable, resp)
		return
	}
	if at.GetSecretFromHeader(req) == prof.Secret {
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			resp.Response = err.Error()
			at.WriteResponse(respw, http.StatusBadRequest, resp)
			return
		} else if msg.Message != "" {
			_, err = atdb.InsertOneDoc(config.Mongoconn, "inbox", msg)
			if err != nil {
				resp.Response = err.Error()
			}
			resp, err = at.WebHook(prof.QRKeyword, waphonenumber, config.WAAPIQRLogin, config.WAAPIMessageText, msg, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
			}
			at.WriteResponse(respw, http.StatusOK, resp)
			return
		} else {
			resp.Response = "pesan kosong"
			at.WriteResponse(respw, http.StatusOK, resp)
			return
		}
	}
	resp.Response = "Wrong Secret"
	at.WriteResponse(respw, http.StatusForbidden, resp)
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	httpstatus := http.StatusServiceUnavailable

	// Membuat filter kosong untuk mengambil semua dokumen
	filter := bson.M{}

	profs, err := atdb.GetAllDoc[[]itmodel.Profile](config.Mongoconn, "profile", filter)
	if err != nil {
		resp.Response = err.Error()
	} else {
		for _, prof := range profs {
			dt := &itmodel.WebHook{
				URL:    prof.URL,
				Secret: prof.Secret,
			}
			res, err := at.RefreshToken(dt, prof.Phonenumber, config.WAAPIGetToken, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
				break
			} else {
				resp.Response = at.Jsonstr(res.ModifiedCount)
				httpstatus = http.StatusOK
			}
		}
	}

	at.WriteResponse(respw, httpstatus, resp)
}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	resp.Response = "Not Found"
	at.WriteResponse(respw, http.StatusNotFound, resp)
}

func GetAppProfile(phonenumber string, db *mongo.Database) (apitoken itmodel.Profile, err error) {
	filter := bson.M{"phonenumber": phonenumber}
	apitoken, err = atdb.GetOneDoc[itmodel.Profile](db, "profile", filter)

	return
}
