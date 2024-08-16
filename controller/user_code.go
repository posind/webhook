package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
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

// // HandleQRCodeScan handles the QR code scan request and interacts with whatsauth for token verification
// func PutTokenDataUser(respw http.ResponseWriter, req *http.Request) {
//     // Decode the token from the request using watoken and the public key
//     payload, err := watoken.Decode(config.PublicKeyWhatsAuth, helper.GetLoginFromHeader(req))
//     if err != nil {
//         var respn model.Response
//         respn.Status = "Error: Token Tidak Valid"
//         respn.Info = helper.GetLoginFromHeader(req)
//         respn.Location = "Decode Token Error: " + helper.GetLoginFromHeader(req)
//         respn.Response = err.Error()
//         helper.WriteJSON(respw, http.StatusForbidden, respn)
//         return
//     }

    // Fetch the user data from the database based on the phone number
    docuser, err := atdb.GetOneDoc[model.Profile_user](config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id})
    if err != nil {
        // If the user is not found, create a new user with the payload data
        docuser.PhoneNumber = payload.Id
        docuser.Email = payload.Alias
        helper.WriteJSON(respw, http.StatusNotFound, docuser)
        return
    }

    // Update the user's name/alias
    docuser.Email = payload.Alias

    // Get QRIS status from the WAAPI using the phone number from the payload
    hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetToken + helper.GetLoginFromHeader(req))
    if err != nil {
        helper.WriteJSON(respw, http.StatusMisdirectedRequest, docuser)
        return
    }

    // If the QRIS status is OK and the QR status is not active, generate a new token
    if hcode == http.StatusOK && !qrstat.Status {
        docuser.Token, err = watoken.EncodeforHours(docuser.PhoneNumber, docuser.Email, config.PrivateKey, 43830)
        if err != nil {
            helper.WriteJSON(respw, http.StatusFailedDependency, docuser)
            return
        }
    } else {
        // If the QR status is active, reset the LinkedDevice
        docuser.Token = ""
    }

    // Replace or update the user's data in the "user" collection
    _, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
    if err != nil {
        helper.WriteJSON(respw, http.StatusExpectationFailed, docuser)
        return
    }

    // Respond with the updated user data
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
	if	helper.GetSecretFromHeader(req) != prof.Secret {
		resp.Response = "Salah secret: " + helper.GetSecretFromHeader(req)
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
