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

// func GetDataUser(respw http.ResponseWriter, req *http.Request) {
//     // Decode the token from the request using watoken and the public key
//     payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
//     if err != nil {
//         var respn model.Response
//         respn.Status = "Error: Token Tidak Valid"
//         respn.Info = at.GetLoginFromHeader(req)
//         respn.Location = "Decode Token Error: " + at.GetLoginFromHeader(req)
//         respn.Response = err.Error()
//         at.WriteJSON(respw, http.StatusForbidden, respn)
//         return
//     }

//     // Fetch the user data from the "user" collection
//     var docuser model.User
//     err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, &docuser)
//     if err != nil {
//         var respn model.Response
//         respn.Status = "User not found"
//         respn.Info = err.Error()
//         at.WriteJSON(respw, http.StatusNotFound, respn)
//         return
//     }

//     // Respond with the user data
//     at.WriteJSON(respw, http.StatusOK, docuser)
// }

// func PutTokenDataUser(respw http.ResponseWriter, req *http.Request) {
//     // Decode the token from the request using watoken and the public key
//     payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
//     if err != nil {
//         var respn model.Response
//         respn.Status = "Error: Token Tidak Valid"
//         respn.Info = at.GetLoginFromHeader(req)
//         respn.Location = "Decode Token Error: " + at.GetLoginFromHeader(req)
//         respn.Response = err.Error()
//         at.WriteJSON(respw, http.StatusForbidden, respn)
//         return
//     }

//     // Fetch the user data from the "user" collection
//     var docuser model.User
//     err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, &docuser)
//     if err != nil {
//         docuser.PhoneNumber = payload.Id
//         docuser.Name = payload.Alias
//         at.WriteJSON(respw, http.StatusNotFound, docuser)
//         return
//     }

//     // Update the user's name/alias
//     docuser.Name = payload.Alias

//     // Fetch the QRIS status from the WAAPI
//     hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetToken + at.GetLoginFromHeader(req))
//     if err != nil {
//         at.WriteJSON(respw, http.StatusMisdirectedRequest, docuser)
//         return
//     }

//     // Generate a new token if QRIS status is not active
//     if hcode == http.StatusOK && !qrstat.Status {
//         docuser.Token, err = watoken.EncodeforHours(docuser.PhoneNumber, docuser.Name, config.PrivateKey, 43830) // 5 years
//         if err != nil {
//             at.WriteJSON(respw, http.StatusFailedDependency, docuser)
//             return
//         }
//     } else {
//         docuser.LinkedDevice = ""
//     }

//     // Update or replace the user's data in the "user" collection
//     _, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, docuser)
//     if err != nil {
//         at.WriteJSON(respw, http.StatusExpectationFailed, docuser)
//         return
//     }

//     // Respond with the updated user data
//     at.WriteJSON(respw, http.StatusOK, docuser)
// }

// func PostDataUser(respw http.ResponseWriter, req *http.Request) {
//     // Decode the token from the request using watoken and the public key
//     payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
//     if err != nil {
//         var respn model.Response
//         respn.Status = "Error: Token Tidak Valid"
//         respn.Info = at.GetSecretFromHeader(req)
//         respn.Location = "Decode Token Error"
//         respn.Response = err.Error()
//         at.WriteJSON(respw, http.StatusForbidden, respn)
//         return
//     }

//     // Parse the user data from the request body
//     var usr model.User
//     err = json.NewDecoder(req.Body).Decode(&usr)
//     if err != nil {
//         var respn model.Response
//         respn.Status = "Error: Body tidak valid"
//         respn.Response = err.Error()
//         at.WriteJSON(respw, http.StatusBadRequest, respn)
//         return
//     }

//     // Check if the user exists in the "user" collection
//     var docuser model.User
//     err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, &docuser)
//     if err != nil {
//         // Insert new user
//         usr.PhoneNumber = payload.Id
//         usr.Name = payload.Alias
//         idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
//         if err != nil {
//             var respn model.Response
//             respn.Status = "Gagal Insert Database"
//             respn.Response = err.Error()
//             at.WriteJSON(respw, http.StatusNotModified, respn)
//             return
//         }
//         user.ID = idusr
//         at.WriteJSON(respw, http.StatusOK, usr)
//         return
//     }

//     // Update the user's details
//     docuser.Name = payload.Alias
//     docuser.Email = usr.Email

//     _, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, docuser)
//     if err != nil {
//         var respn model.Response
//         respn.Status = "Gagal replaceonedoc"
//         respn.Response = err.Error()
//         at.WriteJSON(respw, http.StatusConflict, respn)
//         return
//     }

//     at.WriteJSON(respw, http.StatusOK, docuser)
// }

// func PostDataUserFromWA(respw http.ResponseWriter, req *http.Request) {
//     var resp model.Response

//     // Fetch the application profile for WhatsApp
//     prof, err := whatsauth.GetAppProfile(at.GetParam(req), config.Mongoconn)
//     if err != nil {
//         resp.Response = err.Error()
//         at.WriteJSON(respw, http.StatusBadRequest, resp)
//         return
//     }

//     // Validate the secret from the request
//     if at.GetSecretFromHeader(req) != prof.Secret {
//         resp.Response = "Salah secret: " + at.GetSecretFromHeader(req)
//         at.WriteJSON(respw, http.StatusUnauthorized, resp)
//         return
//     }

//     // Parse the user data from the request body
//     var usr model.User
//     err = json.NewDecoder(req.Body).Decode(&usr)
//     if err != nil {
//         resp.Response = "Error: Body tidak valid"
//         resp.Info = err.Error()
//         at.WriteJSON(respw, http.StatusBadRequest, resp)
//         return
//     }

//     // Check if the user exists in the "user" collection
//     var docuser model.User
//     err = atdb.GetOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": usr.PhoneNumber}, &docuser)
//     if err != nil {
//         // Insert the new user into the database
//         idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
//         if err != nil {
//             resp.Response = "Gagal Insert Database"
//             resp.Info = err.Error()
//             at.WriteJSON(respw, http.StatusNotModified, resp)
//             return
//         }
//         resp.Info = idusr.Hex()
//         at.WriteJSON(respw, http.StatusOK, resp)
//         return
//     }

//     // Update the user's data
//     docuser.Name = usr.Name
//     docuser.Email = usr.Email

//     _, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": usr.PhoneNumber}, docuser)
//     if err != nil {
//         resp.Response = "Gagal replaceonedoc"
//         resp.Info = err.Error()
//         at.WriteJSON(respw, http.StatusConflict, resp)
//         return
//     }

//     resp.Info = docuser.ID.Hex()
//     resp.Info = docuser.Email
//     at.WriteJSON(respw, http.StatusOK, resp)
// }

// // // HandleQRCodeScan handles the QR code scan request and interacts with whatsauth for token verification
// // func PutTokenDataUser(respw http.ResponseWriter, req *http.Request) {
// //     // Decode the token from the request using watoken and the public key
// //     payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
// //     if err != nil {
// //         var respn model.Response
// //         respn.Status = "Error: Token Tidak Valid"
// //         respn.Info = at.GetLoginFromHeader(req)
// //         respn.Location = "Decode Token Error: " + at.GetLoginFromHeader(req)
// //         respn.Response = err.Error()
// //         at.WriteJSON(respw, http.StatusForbidden, respn)
// //         return
// //     }

// // Fetch the user data from the database based on the phone number
// // docuser, err := atdb.GetOneDoc[model.Profile_user](config.Mongoconn, "user_login_token", primitive.M{"phonenumber": payload.Id})
// // if err != nil {
// //     // If the user is not found, create a new user with the payload data
// //     docuser.PhoneNumber = payload.Id
// //     docuser.Email = payload.Alias
// //     at.WriteJSON(respw, http.StatusNotFound, docuser)
// //     return
// // }

// // Update the user's name/alias
// // docuser.Email = payload.Alias

// //     // Get QRIS status from the WAAPI using the phone number from the payload
// //     hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetToken + at.GetLoginFromHeader(req))
// //     if err != nil {
// //         at.WriteJSON(respw, http.StatusMisdirectedRequest, docuser)
// //         return
// //     }

// //     // If the QRIS status is OK and the QR status is not active, generate a new token
// //     if hcode == http.StatusOK && !qrstat.Status {
// //         docuser.Token, err = watoken.EncodeforHours(docuser.PhoneNumber, docuser.Email, config.PrivateKey, 43830)
// //         if err != nil {
// //             at.WriteJSON(respw, http.StatusFailedDependency, docuser)
// //             return
// //         }
// //     } else {
// //         // If the QR status is active, reset the LinkedDevice
// //         docuser.Token = ""
// //     }

// //     // Replace or update the user's data in the "user" collection
// //     _, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
// //     if err != nil {
// //         at.WriteJSON(respw, http.StatusExpectationFailed, docuser)
// //         return
// //     }

// //     // Respond with the updated user data
// //     at.WriteJSON(respw, http.StatusOK, docuser)
// // }

// // // PostDataUser handles the POST request to update user data
// // func PostDataUser(respw http.ResponseWriter, req *http.Request) {
// // 	// Decode the token using QRLogin logic
// // 	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
// // 	if err != nil {
// // 		var respn model.Response
// // 		respn.Status = "Error : Token Tidak Valid"
// // 		respn.Info = at.GetSecretFromHeader(req)
// // 		respn.Location = "Decode Token Error"
// // 		respn.Response = err.Error()
// // 		at.WriteJSON(respw, http.StatusForbidden, respn)
// // 		return
// // 	}

// // 	// Parse the user data from the request body
// // 	var usr model.User
// // 	err = json.NewDecoder(req.Body).Decode(&usr)
// // 	if err != nil {
// // 		var respn model.Response
// // 		respn.Status = "Error : Body tidak valid"
// // 		respn.Response = err.Error()
// // 		at.WriteJSON(respw, http.StatusBadRequest, respn)
// // 		return
// // 	}

// // 	// Check if the user already exists in the database
// // 	docuser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": payload.Id})
// // 	if err != nil {
// // 		usr.PhoneNumber = payload.Id
// // 		usr.Name = payload.Alias
// // 		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
// // 		if err != nil {
// // 			var respn model.Response
// // 			respn.Status = "Gagal Insert Database"
// // 			respn.Response = err.Error()
// // 			at.WriteJSON(respw, http.StatusNotModified, respn)
// // 			return
// // 		}
// // 		usr.ID = idusr
// // 		at.WriteJSON(respw, http.StatusOK, usr)
// // 		return
// // 	}

// // 	// Update the user's details
// // 	docuser.Name = payload.Alias
// // 	docuser.Email = usr.Email
// // 	docuser.GitHostUsername = usr.GitHostUsername
// // 	docuser.GitlabUsername = usr.GitlabUsername
// // 	docuser.GithubUsername = usr.GithubUsername

// // 	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": payload.Id}, docuser)
// // 	if err != nil {
// // 		var respn model.Response
// // 		respn.Status = "Gagal replaceonedoc"
// // 		respn.Response = err.Error()
// // 		at.WriteJSON(respw, http.StatusConflict, respn)
// // 		return
// // 	}

// // 	// Update projects where the user is a member
// // 	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", bson.M{"members._id": docuser.ID})
// // 	if err != nil || len(existingprjs) == 0 {
// // 		at.WriteJSON(respw, http.StatusOK, docuser)
// // 		return
// // 	}

// // 	// Loop through each project and update the user
// // 	for _, prj := range existingprjs {
// // 		memberToDelete := model.User{PhoneNumber: docuser.PhoneNumber}
// // 		_, err := atdb.DeleteDocFromArray[model.User](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
// // 		if err != nil {
// // 			var respn model.Response
// // 			respn.Status = "Error : Data project tidak di temukan"
// // 			respn.Response = err.Error()
// // 			at.WriteJSON(respw, http.StatusNotFound, respn)
// // 			return
// // 		}
// // 		_, err = atdb.AddDocToArray[model.User](config.Mongoconn, "project", prj.ID, "members", docuser)
// // 		if err != nil {
// // 			var respn model.Response
// // 			respn.Status = "Error : Gagal menambahkan member ke project"
// // 			respn.Response = err.Error()
// // 			at.WriteJSON(respw, http.StatusExpectationFailed, respn)
// // 			return
// // 		}
// // 	}

// // 	at.WriteJSON(respw, http.StatusOK, docuser)
// // }

// // func PostDataUserFromWA(respw http.ResponseWriter, req *http.Request) {
// // 	var resp itmodel.Response

// // 	// Fetch the application profile for WhatsApp
// // 	prof, err := whatsauth.GetAppProfile(at.GetParam(req), config.Mongoconn)
// // 	if err != nil {
// // 		resp.Response = err.Error()
// // 		at.WriteJSON(respw, http.StatusBadRequest, resp)
// // 		return
// // 	}

// // 	// Validate the secret from the request
// // 	if at.GetSecretFromHeader(req) != prof.Secret {
// // 		resp.Response = "Salah secret: " + at.GetSecretFromHeader(req)
// // 		at.WriteJSON(respw, http.StatusUnauthorized, resp)
// // 		return
// // 	}

// // 	// Decode the user data from the request body
// // 	var usr model.User
// // 	err = json.NewDecoder(req.Body).Decode(&usr)
// // 	if err != nil {
// // 		resp.Response = "Error : Body tidak valid"
// // 		resp.Info = err.Error()
// // 		at.WriteJSON(respw, http.StatusBadRequest, resp)
// // 		return
// // 	}

// // 	// Check if the user exists in the database
// // 	docuser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": usr.PhoneNumber})
// // 	if err != nil {
// // 		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
// // 		if err != nil {
// // 			resp.Response = "Gagal Insert Database"
// // 			resp.Info = err.Error()
// // 			at.WriteJSON(respw, http.StatusNotModified, resp)
// // 			return
// // 		}
// // 		resp.Info = idusr.Hex()
// // 		at.WriteJSON(respw, http.StatusOK, resp)
// // 		return
// // 	}

// // 	// Update the user's data
// // 	docuser.Name = usr.Name
// // 	docuser.Email = usr.Email

// // 	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": usr.PhoneNumber}, docuser)
// // 	if err != nil {
// // 		resp.Response = "Gagal replaceonedoc"
// // 		resp.Info = err.Error()
// // 		at.WriteJSON(respw, http.StatusConflict, resp)
// // 		return
// // 	}

// // 	// Update user membership in projects (if needed)
// // 	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", bson.M{"members._id": docuser.ID})
// // 	if err != nil || len(existingprjs) == 0 {
// // 		resp.Response = "belum terdaftar di project manapun"
// // 		at.WriteJSON(respw, http.StatusOK, resp)
// // 		return
// // 	}

// // 	// Loop through and update the projects where the user is a member
// // 	for _, prj := range existingprjs {
// // 		memberToDelete := model.User{PhoneNumber: docuser.PhoneNumber}
// // 		_, err := atdb.DeleteDocFromArray[model.User](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
// // 		if err != nil {
// // 			resp.Response = "Error : Data project tidak di temukan"
// // 			resp.Info = err.Error()
// // 			at.WriteJSON(respw, http.StatusNotFound, resp)
// // 			return
// // 		}
// // 		_, err = atdb.AddDocToArray[model.User](config.Mongoconn, "project", prj.ID, "members", docuser)
// // 		if err != nil {
// // 			resp.Response = "Error : Gagal menambahkan member ke project"
// // 			resp.Info = err.Error()
// // 			at.WriteJSON(respw, http.StatusExpectationFailed, resp)
// // 			return
// // 		}
// // 	}

// // 	resp.Info = docuser.ID.Hex()
// // 	resp.Info = docuser.Email
// // 	at.WriteJSON(respw, http.StatusOK, resp)
// // }

