package whatsauth

import (
	"strings"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/kimseok"
	"github.com/gocroot/helper/normalize"

	"github.com/gocroot/mod"

	"github.com/gocroot/helper/module"
	"github.com/whatsauth/itmodel"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func WebHook(profile itmodel.Profile, msg itmodel.IteungMessage, db *mongo.Database) (resp itmodel.Response, err error) {
	if IsLoginRequest(msg, profile.QRKeyword) { //untuk whatsauth request login
		resp, err = HandlerQRLogin(msg, profile, db)
	} else { //untuk membalas pesan masuk
		resp, err = HandlerIncomingMessage(msg, profile, db)
	}
	return
}

func RefreshToken(dt *itmodel.WebHook, WAPhoneNumber, WAAPIGetToken string, db *mongo.Database) (res *mongo.UpdateResult, err error) {
	profile, err := GetAppProfile(WAPhoneNumber, db)
	if err != nil {
		return
	}
	var resp itmodel.User
	if profile.Token != "" {
		resp, err = atapi.PostStructWithToken[itmodel.User]("Token", profile.Token, dt, WAAPIGetToken)
		if err != nil {
			return
		}
		profile.Phonenumber = resp.PhoneNumber
		profile.Token = resp.Token
		res, err = atdb.ReplaceOneDoc(db, "profile", bson.M{"phonenumber": resp.PhoneNumber}, profile)
		if err != nil {
			return
		}
	}
	return
}

func IsLoginRequest(msg itmodel.IteungMessage, keyword string) bool {
	return strings.Contains(msg.Message, keyword) // && msg.From_link
}

func GetUUID(msg itmodel.IteungMessage, keyword string) string {
	return strings.Replace(msg.Message, keyword, "", 1)
}

func HandlerQRLogin(msg itmodel.IteungMessage, profile itmodel.Profile, db *mongo.Database) (resp itmodel.Response, err error) {
	dt := &itmodel.WhatsauthRequest{
		Uuid:        GetUUID(msg, profile.QRKeyword),
		Phonenumber: msg.Phone_number,
		Aliasname:   msg.Alias_name,
		Delay:       msg.From_link_delay,
	}
	structtoken, err := GetAppProfile(profile.Phonenumber, db)
	if err != nil {
		return
	}
	resp, err = atapi.PostStructWithToken[itmodel.Response]("Token", structtoken.Token, dt, profile.URLQRLogin)
	return
}

func HandlerIncomingMessage(msg itmodel.IteungMessage, profile itmodel.Profile, db *mongo.Database) (resp itmodel.Response, err error) {
	_, bukanbot := GetAppProfile(msg.Phone_number, db) //cek apakah nomor adalah bot
	if bukanbot != nil {                               //jika tidak terdapat di profile
		msg.Message = normalize.NormalizeHiddenChar(msg.Message)
		module.NormalizeAndTypoCorrection(&msg.Message, db, "typo")
		modname, group, personal := module.GetModuleName(profile.Phonenumber, msg, db, "module")
		var msgstr string
		var isgrup bool
		if msg.Chat_server != "g.us" { //chat personal
			if personal && modname != "" {
				msgstr = mod.Caller(profile, modname, msg, db)
			} else {
				msgstr = kimseok.GetMessage(profile, msg, profile.Botname, db)
			}

		} else if strings.Contains(strings.ToLower(msg.Message), profile.Triggerword) { //chat group
			msg.Message = HapusNamaPanggilanBot(msg.Message, profile.Triggerword, profile.Botname)
			//set grup true
			isgrup = true
			if group && modname != "" {
				msgstr = mod.Caller(profile, modname, msg, db)
			} else {
				msgstr = kimseok.GetMessage(profile, msg, profile.Botname, db)
			}
		}
		dt := &itmodel.TextMessage{
			To:       msg.Chat_number,
			IsGroup:  isgrup,
			Messages: msgstr,
		}
		resp, err = atapi.PostStructWithToken[itmodel.Response]("Token", profile.Token, dt, profile.URLAPIText)
		if err != nil {
			return
		}

	}
	return
}

// HapusNamaPanggilanBot menghapus semua kemunculan nama panggilan dan nama lengkap dari pesan
func HapusNamaPanggilanBot(msg string, namapanggilan string, namalengkap string) string {
	// Mengubah pesan dan nama panggilan menjadi lowercase untuk pencocokan yang tidak peka huruf besar-kecil
	namapanggilan = strings.ToLower(namapanggilan)
	namalengkap = strings.ToLower(namalengkap)
	msg = strings.ToLower(msg)

	// Hapus semua kemunculan nama lengkap dari pesan
	msg = strings.ReplaceAll(msg, namalengkap+" ", "")
	msg = strings.ReplaceAll(msg, " "+namalengkap, "")
	//msg = strings.ReplaceAll(msg, namalengkap, "")

	// Hapus semua kemunculan nama panggilan dari pesan
	msg = strings.ReplaceAll(msg, namapanggilan+" ", "")
	msg = strings.ReplaceAll(msg, " "+namapanggilan, "")
	//msg = strings.ReplaceAll(msg, namapanggilan, "")

	// Menghapus spasi tambahan jika ada
	msg = strings.TrimSpace(msg)

	return msg
}

func GetRandomReplyFromMongo(msg itmodel.IteungMessage, botname string, db *mongo.Database) string {
	rply, err := atdb.GetRandomDoc[itmodel.Reply](db, "reply", 1)
	if err != nil {
		return "Koneksi Database Gagal: " + err.Error()
	}
	replymsg := strings.ReplaceAll(rply[0].Message, "#BOTNAME#", botname)
	replymsg = strings.ReplaceAll(replymsg, "\\n", "\n")
	return replymsg
}

func GetAppProfile(phonenumber string, db *mongo.Database) (apitoken itmodel.Profile, err error) {
	filter := bson.M{"phonenumber": phonenumber}
	apitoken, err = atdb.GetOneDoc[itmodel.Profile](db, "profile", filter)

	return
}
