package helper

import (
	"strings"

	"github.com/gocroot/mod"

	"github.com/gocroot/helper/module"
	"github.com/whatsauth/itmodel"

	"github.com/gocroot/helper/kimseok"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func WebHook(WAKeyword, WAPhoneNumber, WAAPIQRLogin, WAAPIMessage string, msg itmodel.IteungMessage, db *mongo.Database) (resp itmodel.Response, err error) {
	if IsLoginRequest(msg, WAKeyword) { //untuk whatsauth request login
		resp, err = HandlerQRLogin(msg, WAKeyword, WAPhoneNumber, db, WAAPIQRLogin)
	} else { //untuk membalas pesan masuk
		resp, err = HandlerIncomingMessage(msg, WAPhoneNumber, db, WAAPIMessage)
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
		resp, err = PostStructWithToken[itmodel.User]("Token", profile.Token, dt, WAAPIGetToken)
		if err != nil {
			return
		}
		profile.Phonenumber = resp.PhoneNumber
		profile.Token = resp.Token
		res, err = ReplaceOneDoc(db, "profile", bson.M{"phonenumber": resp.PhoneNumber}, profile)
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

func HandlerQRLogin(msg itmodel.IteungMessage, WAKeyword string, WAPhoneNumber string, db *mongo.Database, WAAPIQRLogin string) (resp itmodel.Response, err error) {
	dt := &itmodel.WhatsauthRequest{
		Uuid:        GetUUID(msg, WAKeyword),
		Phonenumber: msg.Phone_number,
		Aliasname:   msg.Alias_name,
		Delay:       msg.From_link_delay,
	}
	structtoken, err := GetAppProfile(WAPhoneNumber, db)
	if err != nil {
		return
	}
	resp, err = PostStructWithToken[itmodel.Response]("Token", structtoken.Token, dt, WAAPIQRLogin)
	return
}

func HandlerIncomingMessage(msg itmodel.IteungMessage, WAPhoneNumber string, db *mongo.Database, WAAPIMessage string) (resp itmodel.Response, err error) {
	_, bukanbot := GetAppProfile(msg.Phone_number, db) //cek apakah nomor adalah bot
	if bukanbot != nil {                               //jika tidak terdapat di profile
		var profile itmodel.Profile
		profile, err = GetAppProfile(WAPhoneNumber, db)
		if err != nil {
			return
		}
		module.NormalizeAndTypoCorrection(&msg.Message, db, "typo")
		modname, group, personal := module.GetModuleName(WAPhoneNumber, msg, db, "module")
		var msgstr string
		if msg.Chat_server != "g.us" { //chat personal
			if personal && modname != "" {
				msgstr = mod.Caller(profile, modname, msg, db)
			} else {
				msgstr = GetRandomReplyFromMongo(msg, profile.Botname, db)
			}
			//
			if strings.Contains(msgstr, "IM$G#M$Gui76557u|||") {
				strdt := strings.Split(msgstr, "|||")
				dt := &itmodel.ImageMessage{
					To:          msg.Chat_number,
					Base64Image: strdt[1],
					Caption:     strdt[2],
					IsGroup:     false,
				}
				urlpost := "https://api.wa.my.id/api/send/message/image"
				resp, err = PostStructWithToken[itmodel.Response]("Token", profile.Token, dt, urlpost)
				if err != nil {
					return
				}
			} else {
				dt := &itmodel.TextMessage{
					To:       msg.Chat_number,
					IsGroup:  false,
					Messages: msgstr,
				}
				resp, err = PostStructWithToken[itmodel.Response]("Token", profile.Token, dt, WAAPIMessage)
				if err != nil {
					return
				}
			}

		} else if strings.Contains(strings.ToLower(msg.Message), profile.Triggerword) { //chat group
			if group && modname != "" {
				msgstr = mod.Caller(profile, modname, msg, db)
			} else {
				msgstr = GetRandomReplyFromMongo(msg, profile.Botname, db)
			}
			if strings.Contains(msgstr, "IM$G#M$Gui76557u|||") {
				strdt := strings.Split(msgstr, "|||")
				dt := &itmodel.ImageMessage{
					To:          msg.Chat_number,
					Base64Image: strdt[1],
					Caption:     strdt[2],
					IsGroup:     true,
				}
				urlpost := "https://api.wa.my.id/api/send/message/image"
				resp, err = PostStructWithToken[itmodel.Response]("Token", profile.Token, dt, urlpost)
				if err != nil {
					return
				}
			} else {
				dt := &itmodel.TextMessage{
					To:       msg.Chat_number,
					IsGroup:  true,
					Messages: msgstr,
				}
				resp, err = PostStructWithToken[itmodel.Response]("Token", profile.Token, dt, WAAPIMessage)
				if err != nil {
					return
				}
			}

		}

	}
	return
}

func GetRandomReplyFromMongo(msg itmodel.IteungMessage, botname string, db *mongo.Database) string {
	rply, err := GetRandomDoc[itmodel.Reply](db, "reply", 1)
	if err != nil {
		return "Koneksi Database Gagal: " + err.Error()
	}
	//jika tidak ada di db komplain lanjut ke selanjutnya
	if len(rply) == 0 {
		dt, err := kimseok.QueriesDataRegexpALL(db, msg.Message)
		if err != nil {
			return err.Error()
		}
		rply = []itmodel.Reply{{Message: strings.TrimSpace(dt.Answer)}}

	}
	replymsg := strings.ReplaceAll(rply[0].Message, "#BOTNAME#", botname)
	replymsg = strings.ReplaceAll(replymsg, "\\n", "\n")
	return replymsg
}

func GetAppProfile(phonenumber string, db *mongo.Database) (apitoken itmodel.Profile, err error) {
	filter := bson.M{"phonenumber": phonenumber}
	apitoken, err = GetOneDoc[itmodel.Profile](db, "profile", filter)

	return
}
