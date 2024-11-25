package helpdesk

import (
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPrefillMessage(key string, db *mongo.Database) (value string) {
	templ, err := atdb.GetOneDoc[model.Prefill](db, "prefill", bson.M{"key": key})
	if err != nil {
		return
	}
	return templ.Value
}
