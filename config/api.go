package config

import "os"

var WAAPIQRLogin string = "https://api.wa.my.id/api/whatsauth/request"

var WAAPIMessageText string = "https://api.wa.my.id/api/send/message/text"

var WAAPIMessageImage string = "https://api.wa.my.id/api/send/message/image"

var WAAPIGetToken string = "https://api.wa.my.id/api/signup"

var PublicKeyWhatsAuth string

var WAAPIToken string

var APIGETPDLMS string = "https://pamongdesa.kemendagri.go.id/webservice/public/user/get-by-phone?number="

var APITOKENPD string = os.Getenv("PDTOKEN")
