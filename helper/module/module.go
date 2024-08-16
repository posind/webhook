package module

import (
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetModuleName(WAPhoneNumber string, im itmodel.IteungMessage, MongoConn *mongo.Database, ModuleCollection string) (modulename string, group bool, personal bool) {
	modules, _ := atdb.GetAllDoc[[]Module](MongoConn, ModuleCollection, bson.M{"phonenumbers": WAPhoneNumber})
	for _, mod := range modules {
		complete, _ := IsMatch(strings.ToLower(im.Message), mod.Keyword...)
		if complete {
			modulename = mod.Name
			group = mod.Group
			personal = mod.Personal
		}
	}
	return
}

func IsMatch(str string, subs ...string) (bool, int) {

	matches := 0
	isCompleteMatch := true
	for _, sub := range subs {
		if strings.Contains(str, sub) {
			matches += 1
		} else {
			isCompleteMatch = false
		}
	}

	return isCompleteMatch, matches
}
