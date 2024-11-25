package lms

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RefreshCookie(db *mongo.Database) (err error) {
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	newxs, newls, newbar, err := GetNewCookie(profile.Xsrf, profile.Lsession, db)
	if err != nil {
		return
	}
	profile.Bearer = newbar
	profile.Xsrf = newxs
	profile.Lsession = newls
	_, err = atdb.ReplaceOneDoc(db, "lmscreds", bson.M{"username": "madep"}, profile)
	if err != nil {
		return
	}
	return

}

func GetTotalUser(db *mongo.Database) (total int, err error) {
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	url := profile.URLUsers
	url = strings.ReplaceAll(url, "##TOTAL##", "1")

	_, res, err := atapi.GetWithBearer[Root](profile.Bearer, url)
	if err != nil {
		err = errors.New("GetWithBearer:" + err.Error() + " " + url + " " + profile.Bearer)
		return
	}
	total = res.Data.Meta.Total
	return
}

func GetAllUser(db *mongo.Database) (users []User, err error) {
	total, err := GetTotalUser(db)
	if err != nil {
		err = errors.New("GetTotalUser:" + err.Error())
		return
	}
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	url := profile.URLUsers
	url = strings.ReplaceAll(url, "##TOTAL##", strconv.Itoa(total))
	_, res, err := atapi.GetWithBearer[Root](profile.Bearer, url)
	if err != nil {
		err = errors.New("GetWithBearer:" + err.Error() + profile.Bearer + " " + url)
		return
	}
	users = res.Data.Data
	return
}
