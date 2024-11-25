package helpdesk

import (
	"context"
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// mendapatkan nama team helpdesk dari pesan
func GetNamaTeamFromPesan(Pesan itmodel.IteungMessage, db *mongo.Database) (team string, helpdeskslist []string, err error) {
	msg := strings.ReplaceAll(Pesan.Message, "bantuan", "")
	msg = strings.ReplaceAll(msg, "operator", "")
	msg = strings.TrimSpace(msg)
	//ambil dulu semua nama team di database
	helpdesks, err := atdb.GetAllDistinctDoc(db, bson.M{}, "team", "user")
	if err != nil {
		return
	}
	//pecah kalimat batasan spasi
	msgs := strings.Fields(msg)
	//jika nama team tidak ada atau hanya kata bantuan operator saja, maka keluarkan list nya
	if len(msgs) != 0 {
		msg = msgs[0]
	}
	//mendapatkan keyword dari kata pertama dalam kalimat masuk ke team yang mana
	for _, helpdesk := range helpdesks {
		tim := helpdesk.(string)
		if strings.EqualFold(msg, tim) {
			team = tim
			return
		}
		helpdeskslist = append(helpdeskslist, tim)
	}
	return
}

// mendapatkan scope helpdesk dari pesan
func GetScopeFromTeam(Pesan itmodel.IteungMessage, team string, db *mongo.Database) (scope string, scopeslist []string, err error) {
	msg := strings.ReplaceAll(Pesan.Message, "bantuan", "")
	msg = strings.ReplaceAll(msg, "operator", "")
	msg = strings.ReplaceAll(msg, team, "")
	msg = strings.TrimSpace(msg)
	filter := bson.M{
		"team": team,
	}
	//ambil dulu semua scope di db berdasarkan team
	scopes, err := atdb.GetAllDistinctDoc(db, filter, "scope", "user")
	if err != nil {
		return
	}
	//mendapatkan keyword masuk ke team yang mana
	for _, scp := range scopes {
		scpe := scp.(string)
		if strings.EqualFold(msg, scpe) {
			scope = scpe
			return
		}
		scopeslist = append(scopeslist, scpe)
	}
	return
}

// mendapatkan scope helpdesk dari pesan
func GetOperatorFromScopeandTeam(scope, team string, db *mongo.Database) (operator model.Userdomyikado, err error) {
	filter := bson.M{
		"scope": scope,
		"team":  team,
	}
	operator, err = atdb.GetOneLowestDoc[model.Userdomyikado](db, "user", filter, "jumlahantrian")
	if err != nil {
		return
	}
	operator.JumlahAntrian += 1
	filter = bson.M{
		"scope":       scope,
		"team":        team,
		"phonenumber": operator.PhoneNumber,
	}
	_, err = atdb.ReplaceOneDoc(db, "user", filter, operator)
	if err != nil {
		return
	}
	return
}

// mendapatkan section helpdesk dari pesan
func GetOperatorFromSection(section string, db *mongo.Database) (operator model.Userdomyikado, err error) {
	filter := bson.M{
		"section": section,
	}
	operator, err = atdb.GetOneLowestDoc[model.Userdomyikado](db, "user", filter, "jumlahantrian")
	if err != nil {
		return
	}
	operator.JumlahAntrian += 1
	filter = bson.M{
		"section":     section,
		"phonenumber": operator.PhoneNumber,
	}
	_, err = atdb.ReplaceOneDoc(db, "user", filter, operator)
	if err != nil {
		return
	}
	return
}

func GetSectionFromProvinsiRegex(db *mongo.Database, queries string) (section string, err error) {
	var user model.Userdomyikado
	filter := bson.M{"scope": primitive.Regex{Pattern: queries, Options: "i"}}
	err = db.Collection("datasets").FindOne(context.TODO(), filter).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	section = user.Section
	return
}
