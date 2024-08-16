package module

import (
	"regexp"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NormalizeAndTypoCorrection(message *string, MongoConn *mongo.Database, TypoCollection string) {
	typos, _ := atdb.GetAllDoc[[]Typo](MongoConn, TypoCollection, bson.M{})
	for _, typo := range typos {
		re := regexp.MustCompile(`(?i)` + typo.From + ``)
		*message = re.ReplaceAllString(*message, typo.To)
	}
}
