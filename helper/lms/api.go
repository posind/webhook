package lms

import (
	"github.com/gocroot/config"
	"github.com/gocroot/helper/atapi"
)

func GetNamadanDesaFromAPI(phonenumber string) (namadandesa string) {
	statuscode, res, err := atapi.GetStructWithToken[ResponseAPIPD]("token", config.APITOKENPD, config.APIGETPDLMS+phonenumber)
	if err != nil {
		return
	}
	if statuscode != 200 { //404 jika user not found
		return
	}
	if res.Data.Village != "" {
		namadandesa = res.Data.Fullname + " dari " + res.Data.Village
	} else {
		namadandesa = res.Data.Fullname
	}

	return
}

func GetDataFromAPI(phonenumber string) (data ResponseAPIPD) {
	statuscode, res, err := atapi.GetStructWithToken[ResponseAPIPD]("token", config.APITOKENPD, config.APIGETPDLMS+phonenumber)
	if err != nil {
		return
	}
	if statuscode != 200 { //404 jika user not found
		return
	}
	return res
}
