package controller

// import (
// 	"encoding/json"
// 	"net/http"

// 	"github.com/gocroot/config"
// 	"github.com/gocroot/helper/at"
// 	"github.com/gocroot/helper/atapi"
// 	"github.com/gocroot/helper/atdb"
// 	"github.com/gocroot/helper/whatsauth"
// 	"github.com/gocroot/model"
// 	"github.com/whatsauth/watoken"
// 	"go.mongodb.org/mongo-driver/bson"
// )

// // GetDataUserFromApi retrieves user data from an external API based on token
// func GetDataUserFromApi(respw http.ResponseWriter, req *http.Request) {
// 	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Invalid Token", http.StatusForbidden, err, at.GetLoginFromHeader(req))
// 		return
// 	}

// 	// Fetch user data from external API
// 	userdt := lms.GetDataFromAPI(payload.Id)
// 	if userdt.Data.Fullname == "" {
// 		at.WriteJSON(respw, http.StatusNotFound, userdt)
// 		return
// 	}

// 	at.WriteJSON(respw, http.StatusOK, userdt)
// }

// // GetDataUser retrieves user data from MongoDB based on token
// func GetDataUser(respw http.ResponseWriter, req *http.Request) {
// 	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Invalid Token", http.StatusForbidden, err, at.GetLoginFromHeader(req))
// 		return
// 	}

// 	// Retrieve user data from MongoDB
// 	var docuser model.Userdomyikado
// 	err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, &docuser)
// 	if err != nil {
// 		docuser.PhoneNumber = payload.Id
// 		docuser.Name = payload.Alias
// 		at.WriteJSON(respw, http.StatusNotFound, docuser)
// 		return
// 	}

// 	docuser.Name = payload.Alias
// 	at.WriteJSON(respw, http.StatusOK, docuser)
// }

// // PutTokenDataUser checks device linking status and generates a 5-year token if linked
// func PutTokenDataUser(respw http.ResponseWriter, req *http.Request) {
// 	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Invalid Token", http.StatusForbidden, err, at.GetLoginFromHeader(req))
// 		return
// 	}

// 	// Retrieve user data
// 	var docuser model.Userdomyikado
// 	err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, &docuser)
// 	if err != nil {
// 		docuser.PhoneNumber = payload.Id
// 		docuser.Name = payload.Alias
// 		at.WriteJSON(respw, http.StatusNotFound, docuser)
// 		return
// 	}

// 	// Check QRIS device linking status
// 	hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetDevice + at.GetLoginFromHeader(req))
// 	if err != nil {
// 		at.WriteJSON(respw, http.StatusMisdirectedRequest, docuser)
// 		return
// 	}

// 	// Generate 5-year token if not linked
// 	if hcode == http.StatusOK && !qrstat.Status {
// 		docuser.LinkedDevice, err = watoken.EncodeforHours(docuser.PhoneNumber, docuser.Name, config.PrivateKey, 43830)
// 		if err != nil {
// 			at.WriteJSON(respw, http.StatusFailedDependency, docuser)
// 			return
// 		}
// 	} else {
// 		docuser.LinkedDevice = ""
// 	}

// 	// Update user data
// 	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, docuser)
// 	if err != nil {
// 		at.WriteJSON(respw, http.StatusExpectationFailed, docuser)
// 		return
// 	}

// 	at.WriteJSON(respw, http.StatusOK, docuser)
// }

// // PostDataUser inserts or updates user data in MongoDB
// func PostDataUser(respw http.ResponseWriter, req *http.Request) {
// 	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Invalid Token", http.StatusForbidden, err, at.GetSecretFromHeader(req))
// 		return
// 	}

// 	var usr model.Userdomyikado
// 	err = json.NewDecoder(req.Body).Decode(&usr)
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Invalid Request Body", http.StatusBadRequest, err, "")
// 		return
// 	}

// 	// Retrieve user data from MongoDB
// 	var docuser model.Userdomyikado
// 	err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, &docuser)
// 	if err != nil {
// 		usr.PhoneNumber = payload.Id
// 		usr.Name = payload.Alias
// 		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
// 		if err != nil {
// 			sendErrorResponse(respw, "Error: Failed to Insert into Database", http.StatusNotModified, err, "")
// 			return
// 		}
// 		usr.ID = idusr
// 		at.WriteJSON(respw, http.StatusOK, usr)
// 		return
// 	}

// 	// Update existing user data
// 	docuser.Name = payload.Alias
// 	docuser.Email = usr.Email
// 	docuser.GitHostUsername = usr.GitHostUsername
// 	docuser.GitlabUsername = usr.GitlabUsername
// 	docuser.GithubUsername = usr.GithubUsername

// 	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, docuser)
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Failed to Replace User Document", http.StatusConflict, err, "")
// 		return
// 	}

// 	// Update membership in projects
// 	updateProjectsMembership(respw, docuser)
// }

// // PostDataUserFromWA processes WhatsApp user data and updates MongoDB
// func PostDataUserFromWA(respw http.ResponseWriter, req *http.Request) {
// 	var resp model.Response
// 	prof, err := whatsauth.GetAppProfile(at.GetParam(req), config.Mongoconn)
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Failed to Get App Profile", http.StatusBadRequest, err, "")
// 		return
// 	}

// 	// Validate secret from the request header
// 	if at.GetSecretFromHeader(req) != prof.Secret {
// 		sendErrorResponse(respw, "Error: Invalid Secret", http.StatusUnauthorized, nil, "")
// 		return
// 	}

// 	// Decode incoming request body
// 	var usr model.Userdomyikado
// 	err = json.NewDecoder(req.Body).Decode(&usr)
// 	if err != nil {
// 		sendErrorResponse(respw, "Error: Invalid Request Body", http.StatusBadRequest, err, "")
// 		return
// 	}

// 	// Insert or update user data in MongoDB
// 	processUserData(respw, usr)
// }

// // Helper function to process user data insertion and updating in MongoDB
// func processUserData(respw http.ResponseWriter, usr model.Userdomyikado) {
// 	var resp model.Response
// 	var docuser model.Userdomyikado

// 	err := atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": usr.PhoneNumber}, &docuser)
// 	if err != nil {
// 		// Insert new user
// 		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
// 		if err != nil {
// 			resp.Response = "Error: Failed to Insert into Database"
// 			resp.Info = err.Error()
// 			at.WriteJSON(respw, http.StatusNotModified, resp)
// 			return
// 		}
// 		resp.Info = idusr.Hex()
// 		at.WriteJSON(respw, http.StatusOK, resp)
// 		return
// 	}

// 	// Update user details
// 	docuser.Name = usr.Name
// 	docuser.Email = usr.Email
// 	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": usr.PhoneNumber}, docuser)
// 	if err != nil {
// 		resp.Response = "Error: Failed to Replace User Document"
// 		resp.Info = err.Error()
// 		at.WriteJSON(respw, http.StatusConflict, resp)
// 		return
// 	}

// 	// Update project memberships
// 	updateProjectsMembership(respw, docuser)
// }

// // Helper function to update project memberships for the user
// func updateProjectsMembership(respw http.ResponseWriter, docuser model.Userdomyikado) {
// 	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", bson.M{"members._id": docuser.ID})
// 	if err != nil || len(existingprjs) == 0 {
// 		at.WriteJSON(respw, http.StatusOK, docuser)
// 		return
// 	}

// 	for _, prj := range existingprjs {
// 		// Remove old member data and add updated data
// 		memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
// 		_, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
// 		if err != nil {
// 			sendErrorResponse(respw, "Error: Failed to Update Project Membership", http.StatusNotFound, err, "")
// 			return
// 		}

// 		_, err = atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", docuser)
// 		if err != nil {
// 			sendErrorResponse(respw, "Error: Failed to Add Member to Project", http.StatusExpectationFailed, err, "")
// 			return
// 		}
// 	}

// 	at.WriteJSON(respw, http.StatusOK, docuser)
// }

// // Helper function to send error responses
// func sendErrorResponse(w http.ResponseWriter, message string, statusCode int, err error, info string) {
// 	var resp model.Response
// 	resp.Status = "Error"
// 	resp.Response = message
// 	if err != nil {
// 		resp.Info = err.Error()
// 	} else {
// 		resp.Info = info
// 	}
// 	at.WriteJSON(w, statusCode, resp)
// }
