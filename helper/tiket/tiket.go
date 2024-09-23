package tiket

import (
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/helper/waktu"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// phonefield: userphone or adminphone
func IsTicketClosed(phonefield string, phonenumber string, db *mongo.Database) (closed bool, stiket Bantuan, err error) {
	stiket, err = atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, phonefield: phonenumber})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//tiket udah close karena no doc
			closed = true
			err = nil // Reset err ke nil karena ini bukan error, hanya kondisi normal
			return
		}
		// Jika ada error lain, kita return error tersebut
		return
	}
	// Jika ada tiket yang belum closed, kita kembalikan nilai default closed = false
	return
}

func InserNewTicket(userphone string, adminname string, adminphone string, db *mongo.Database) (IDTiket primitive.ObjectID, err error) {
	dataapi := lms.GetDataFromAPI(userphone)
	tiketbaru := Bantuan{
		UserName:   dataapi.Data.Fullname,
		UserPhone:  userphone,
		AdminPhone: adminphone,
		AdminName:  adminname,
		Prov:       dataapi.Data.Province,
		KabKot:     dataapi.Data.Regency,
		Kec:        dataapi.Data.District,
		Desa:       dataapi.Data.Village,
		StartAt:    waktu.Sekarang(),
	}
	IDTiket, err = atdb.InsertOneDoc(db, "tiket", tiketbaru)
	if err != nil {
		return
	}
	return
}

func UpdateUserMsgInTiket(userphone string, usermsg string, db *mongo.Database) (err error) {
	tiket, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, "userphone": userphone})
	if err != nil {
		return
	}
	wkt, err := waktu.GetDateTimeJKTNow()
	if err != nil {
		return
	}

	tiket.UserMessage += "\n" + wkt + " : " + usermsg
	_, err = atdb.ReplaceOneDoc(db, "tiket", bson.M{"_id": tiket.ID}, tiket)
	if err != nil {
		return
	}
	return
}

func GetNamaAdmin(adminphone string, db *mongo.Database) (name string) {
	tiket, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"adminphone": adminphone})
	if err != nil {
		return
	}
	return tiket.AdminName
}

func UpdateAdminMsgInTiket(adminphone string, adminmsg string, db *mongo.Database) (err error) {
	tiket, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, "adminphone": adminphone})
	if err != nil {
		return
	}
	wkt, err := waktu.GetDateTimeJKTNow()
	if err != nil {
		return
	}
	if tiket.ResponsAt.IsZero() {
		tiket.ResponsAt = waktu.Sekarang()
	}
	tiket.AdminMessage += "\n" + wkt + " : " + adminmsg
	_, err = atdb.ReplaceOneDoc(db, "tiket", bson.M{"_id": tiket.ID}, tiket)
	if err != nil {
		return
	}
	return
}

func IsAdmin(adminphone string, db *mongo.Database) (isadmin bool) {
	_, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"adminphone": adminphone})
	if err != nil {
		return
	}
	isadmin = true
	return
}
