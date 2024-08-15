package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetDataUser handles the GET request to fetch user data
func GetDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKey, helper.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = helper.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error: " + helper.GetLoginFromHeader(req)
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}

	docuser, err := atdb.GetOneDoc[model.Profile_user](config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		helper.WriteJSON(respw, http.StatusNotFound, docuser)
		return
	}

	docuser.Name = payload.Alias
	helper.WriteJSON(respw, http.StatusOK, docuser)
}

// PutTokenDataUser handles the PUT request to update user token data
func PutTokenDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKey, helper.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = helper.GetLoginFromHeader(req)
		respn.Location = "Decode Token Error: " + helper.GetLoginFromHeader(req)
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}

	docuser, err := atdb.GetOneDoc[model.Profile_user](config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Email = payload.Alias
		docuser.CreatedAt = time.Now()
		docuser.UpdatedAt = time.Now()
		helper.WriteJSON(respw, http.StatusNotFound, docuser)
		return
	}

	docuser.Email = payload.Alias
	docuser.UpdatedAt = time.Now()

	hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetToken + helper.GetLoginFromHeader(req))
	if err != nil {
		helper.WriteJSON(respw, http.StatusMisdirectedRequest, docuser)
		return
	}

	if hcode == http.StatusOK && !qrstat.Status {
		docuser.Token, err = watoken.EncodeforHours(docuser.PhoneNumber, docuser.Email, config.PrivateKey, 43830)
		if err != nil {
			helper.WriteJSON(respw, http.StatusFailedDependency, docuser)
			return
		}
	} else {
		docuser.Token = ""
	}

	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id}, docuser)
	if err != nil {
		helper.WriteJSON(respw, http.StatusExpectationFailed, docuser)
		return
	}

	helper.WriteJSON(respw, http.StatusOK, docuser)
}

// PostDataUser handles the POST request to update user data
func PostDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKey, helper.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = helper.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}

	var usr model.Profile_user
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Body Tidak Valid"
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}

	docuser, err := atdb.GetOneDoc[model.Profile_user](config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		usr.PhoneNumber = payload.Id
		usr.Name = payload.Alias
		usr.CreatedAt = time.Now()
		usr.UpdatedAt = time.Now()
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user_login_token", usr)
		if err != nil {
			var respn model.Response
			respn.Status = "Gagal Insert Database"
			respn.Response = err.Error()
			helper.WriteJSON(respw, http.StatusNotModified, respn)
			return
		}
		usr.ID = idusr.(primitive.ObjectID)
		helper.WriteJSON(respw, http.StatusOK, usr)
		return
	}

	docuser.Name = payload.Alias
	docuser.Email = usr.Email
	docuser.UpdatedAt = time.Now()
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id}, docuser)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal ReplaceOneDoc"
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusConflict, respn)
		return
	}

	helper.WriteJSON(respw, http.StatusOK, docuser)
}

// PostDataUserFromWA handles the POST request to update user data from WhatsApp
func PostDataUserFromWA(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	prof, err := helper.GetAppProfile(helper.GetParam(req), config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}

	if helper.GetSecretFromHeader(req) != prof.Secret {
		resp.Response = "Salah Secret: " + helper.GetSecretFromHeader(req)
		helper.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}

	var usr model.Profile_user
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		resp.Response = "Error: Body Tidak Valid"
		resp.Info = err.Error()
		helper.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}

	docuser, err := atdb.GetOneDoc[model.Profile_user](config.Mongoconn, "user_login_token", primitive.M{"phonenumber": usr.PhoneNumber})
	if err != nil {
		usr.CreatedAt = time.Now()
		usr.UpdatedAt = time.Now()
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user_login_token", usr)
		if err != nil {
			resp.Response = "Gagal Insert Database"
			resp.Info = err.Error()
			helper.WriteJSON(respw, http.StatusNotModified, resp)
			return
		}
		if oid, ok := idusr.(primitive.ObjectID); ok {
			resp.Info = oid.Hex()
		} else {
			resp.Info = ""
		}
		helper.WriteJSON(respw, http.StatusOK, resp)
		return
	}

	docuser.Name = usr.Name
	docuser.Email = usr.Email
	docuser.UpdatedAt = time.Now()
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user_login_token", primitive.M{"phonenumber": usr.PhoneNumber}, docuser)
	if err != nil {
		resp.Response = "Gagal ReplaceOneDoc"
		resp.Info = err.Error()
		helper.WriteJSON(respw, http.StatusConflict, resp)
		return
	}

	resp.Status = "Success"
	resp.Info = docuser.ID.Hex()
	resp.Info = docuser.Email
	helper.WriteJSON(respw, http.StatusOK, resp)
}
