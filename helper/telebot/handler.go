package telebot

import (
	"strconv"
	"strings"

	"github.com/gocroot/helper/module"
	"github.com/gocroot/helper/kimseok"
	"github.com/gocroot/mod"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandlerIncomingMessage(msg itmodel.IteungMessage, profile itmodel.Profile, db *mongo.Database) (resp itmodel.Response, err error) {
	module.NormalizeAndTypoCorrection(&msg.Message, db, "typo")
	modname, group, personal := module.GetModuleName(profile.Phonenumber, msg, db, "module")
	var primaryMsg string
	if !msg.Is_group { //chat personal
		if personal && modname != "" {
			primaryMsg = mod.Caller(profile, modname, msg, db)
		} else {
			primaryMsg = kimseok.GetMessageTele(profile, msg, profile.Botname, db)
		}
		//
		if strings.Contains(primaryMsg, "IM$G#M$Gui76557u|||") {
			strdt := strings.Split(primaryMsg, "|||")
			var chatID int64
			chatID, err = strconv.ParseInt(msg.Chat_number, 10, 64)
			if err != nil {
				resp.Response = err.Error()
				resp.Info = "Error converting string to int64"
				return
			}
			if err = SendImageMessage(chatID, strdt[1], strdt[2], profile.TelegramToken); err != nil {
				resp.Response = err.Error()
				return
			}
		} else {
			var chatID int
			chatID, err = strconv.Atoi(msg.Chat_number)
			if err != nil {
				resp.Response = err.Error()
				resp.Info = "Error converting string to int64"
				return
			}
			if err = SendTextMessage(chatID, primaryMsg, profile.TelegramToken); err != nil {
				resp.Response = err.Error()
				return
			}

		}

	} else if strings.Contains(strings.ToLower(msg.Message), profile.Triggerword) { //chat group
		if group && modname != "" {
			primaryMsg = mod.Caller(profile, modname, msg, db)
		} else {
			primaryMsg = kimseok.GetMessageTele(profile, msg, profile.Botname, db)
		}
		if strings.Contains(primaryMsg, "IM$G#M$Gui76557u|||") {
			strdt := strings.Split(primaryMsg, "|||")
			var chatID int64
			chatID, err = strconv.ParseInt(msg.Chat_number, 10, 64)
			if err != nil {
				resp.Response = err.Error()
				resp.Info = "Error converting string to int64"
				return
			}
			if err = SendImageMessage(chatID, strdt[1], strdt[2], profile.TelegramToken); err != nil {
				resp.Response = err.Error()
				return
			}
		} else {
			var chatID int
			chatID, err = strconv.Atoi(msg.Chat_number)
			if err != nil {
				resp.Response = err.Error()
				resp.Info = "Error converting string to int64"
				return
			}
			if err = SendTextMessage(chatID, primaryMsg, profile.TelegramToken); err != nil {
				resp.Response = err.Error()
				return
			}
		}

	}

	return
}
