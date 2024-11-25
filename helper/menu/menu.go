package menu

import (
	"context"
	"strconv"
	"time"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/tiket"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mengecek apakah pesan adalah nomor menu, jika nomor menu maka akan mengubahnya menjadi keyword
func MenuSessionHandler(msg *itmodel.IteungMessage, db *mongo.Database) string {
	//check apakah nomor adalah admin atau user untuk menentukan startmenu
	var startmenu string
	if !tiket.IsAdmin(msg.Phone_number, db) {
		startmenu = "menu"
	} else {
		startmenu = "adminmenu"
	}
	//check apakah ada session, klo ga ada insert sesssion baru
	Sesdoc, ses, err := CheckSession(msg.Phone_number, db)
	if err != nil {
		return err.Error()
	}
	if !ses { //jika tidak ada session atau session=false maka return menu utama user dan update session isi list nomor menunya
		reply, err := GetMenuFromKeywordAndSetSession(startmenu, Sesdoc, db)
		if err != nil {
			return err.Error()
		}
		return reply

	}
	//jika ada session maka cek menu
	//check apakah pesan integer
	menuno, err := strconv.Atoi(msg.Message)
	if err == nil { //kalo pesan adalah nomor
		for _, menu := range Sesdoc.Menulist { //loping di menu list dari session
			if menuno == menu.No { //jika nomor menu sama dengan nomor yang ada di pesan
				reply, err := GetMenuFromKeywordAndSetSession(menu.Keyword, Sesdoc, db) //check apakah ada menu dengan keyword dari nomor menu
				if err != nil {
					//jika di collection menu tidak ada menu dengan keyword tersebut maka kita kembalikan keyword tersebut untuk di proses ke langkah selanjutnya
					msg.Message = menu.Keyword
					return ""
				}
				return reply
			}
		}
		return "Mohon maaf nomor menu yang anda masukkan tidak ada di daftar menu"
	}
	//kalo pesan bukan nomor return kosong
	return ""
}

// check session udah ada atau belum kalo sudah ada maka refresh session
func CheckSession(phonenumber string, db *mongo.Database) (session Session, result bool, err error) {
	session, err = atdb.GetOneDoc[Session](db, "session", bson.M{"phonenumber": phonenumber})
	session.CreatedAt = time.Now()
	session.PhoneNumber = phonenumber
	if err != nil { //insert session klo belum ada
		_, err = db.Collection("session").InsertOne(context.TODO(), session)
		if err != nil {
			return
		}
	} else { //jika sesssion udah ada
		//refresh waktu session dengan waktu sekarang
		_, err = atdb.DeleteManyDocs(db, "session", bson.M{"phonenumber": phonenumber})
		if err != nil {
			return
		}
		_, err = db.Collection("session").InsertOne(context.TODO(), session)
		if err != nil {
			return
		}
		result = true
	}
	return
}

func GetMenuFromKeywordAndSetSession(keyword string, session Session, db *mongo.Database) (msg string, err error) {
	dt, err := atdb.GetOneDoc[Menu](db, "menu", bson.M{"keyword": keyword})
	if err != nil {
		return
	}
	atdb.UpdateOneDoc(db, "session", bson.M{"phonenumber": session.PhoneNumber}, bson.M{"list": dt.List})
	msg = dt.Header + "\n"
	for _, item := range dt.List {
		msg += strconv.Itoa(item.No) + ". " + item.Konten + "\n"
	}
	msg += dt.Footer
	return
}

func InjectSessionMenu(menulist []MenuList, phonenumber string, db *mongo.Database) error {
	_, err := atdb.UpdateOneDoc(db, "session", bson.M{"phonenumber": phonenumber}, bson.M{"list": menulist})
	return err
}
