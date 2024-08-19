package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/telebot"
	"github.com/gocroot/helper/whatsauth"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
)

func TelebotWebhook(w http.ResponseWriter, r *http.Request) {
	var resp itmodel.Response
	waphonenumber := at.GetParam(r)

	var update telebot.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		resp.Response = err.Error()
		at.WriteResponse(w, http.StatusBadRequest, resp)
		return
	}

	chatID := update.Message.Chat.ID
	prof, err := whatsauth.GetAppProfile(waphonenumber, config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteResponse(w, http.StatusServiceUnavailable, resp)
		return
	}

	if update.Message.Contact != nil && update.Message.Contact.PhoneNumber != "" {
		text := "Hello, " + update.Message.From.FirstName + " nomor handphone " + update.Message.Contact.PhoneNumber + " disimpan"
		if err := telebot.SendTextMessage(chatID, text, prof.TelegramToken); err != nil {
			resp.Response = err.Error()
			at.WriteResponse(w, http.StatusConflict, resp)
			return
		}
		_, err := atdb.InsertOneDoc(config.Mongoconn, "teleuser", update)
		if err != nil {
			resp.Response = err.Error()
			at.WriteResponse(w, http.StatusEarlyHints, resp)
			return
		}
	} else {
		updt, err := atdb.GetOneLatestDoc[telebot.Update](config.Mongoconn, "teleuser", bson.M{"message.from.id": update.Message.From.ID})
		if err != nil {
			err := telebot.RequestPhoneNumber(chatID, prof.TelegramToken)
			if err != nil {
				resp.Response = err.Error()
				at.WriteResponse(w, http.StatusExpectationFailed, resp)
				return
			}
		}
		update.Message.Contact = updt.Message.Contact
		//handler message
		if !update.Message.From.IsBot {
			_, err := atdb.InsertOneDoc(config.Mongoconn, "logtele", update)
			if err != nil {
				resp.Response = err.Error()
				at.WriteResponse(w, http.StatusExpectationFailed, resp)
				return
			}
			msg := telebot.ParseUpdateToIteungMessage(update, prof.TelegramToken)
			if msg.Message != "" {
				telebot.HandlerIncomingMessage(msg, prof, config.Mongoconn)
			}

		}
	}

	at.WriteResponse(w, http.StatusOK, resp)
}
