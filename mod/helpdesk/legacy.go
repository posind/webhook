package helpdesk

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// legacy
// handling key word, keyword :bantuan operator
func StartHelpdesk(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	//check apakah tiket dari user sudah di tutup atau belum
	user, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"terlayani": bson.M{"$exists": false}, "phone": Pesan.Phone_number})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return err.Error()
		}
		//berarti tiket udah close semua
	} else { //ada tiket yang belum close
		msgstr := "*Permintaan bantuan dari Pengguna " + user.Nama + " (" + user.Phone + ")*\n\nMohon dapat segera menghubungi beliau melalui WhatsApp di nomor wa.me/" + user.Phone + " untuk memberikan solusi terkait masalah yang sedang dialami:\n\n" + user.Masalah
		msgstr += "\n\nSetelah masalah teratasi, dimohon untuk menginputkan solusi yang telah diberikan ke dalam sistem melalui tautan berikut:\nwa.me/" + Profile.Phonenumber + "?text=" + user.ID.Hex() + "|+solusi+dari+operator+helpdesk+:+"
		dt := &itmodel.TextMessage{
			To:       user.User.PhoneNumber,
			IsGroup:  false,
			Messages: msgstr,
		}
		go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)
		reply = "Segera, Bapak/Ibu akan dihubungkan dengan salah satu Admin kami, *" + user.User.Name + "*.\n\n Mohon tunggu sebentar, kami akan menghubungi Anda melalui WhatsApp di nomor wa.me/" + user.User.PhoneNumber + "\nTerima kasih atas kesabaran Bapak/Ibu"
		//reply = "Kakak kami hubungkan dengan operator kami yang bernama *" + user.User.Name + "* di nomor wa.me/" + user.User.PhoneNumber + "\nMohon tunggu sebentar kami akan kontak kakak melalui nomor tersebut.\n_Terima kasih_"
		return
	}
	//mendapatkan semua nama team dari db
	namateam, helpdeskslist, err := GetNamaTeamFromPesan(Pesan, db)
	if err != nil {
		return err.Error()
	}

	//suruh pilih nama team kalo tidak ada
	if namateam == "" {
		reply = "Selamat datang Bapak/Ibu " + Pesan.Alias_name + "\n\nTerima kasih telah menghubungi kami *Helpdesk LMS Pamong Desa*\n\n"
		reply += "Untuk mendapatkan layanan yang lebih baik, mohon bantuan Bapak/Ibu *untuk memilih regional* tujuan Anda terlebih dahulu:\n"
		for i, helpdesk := range helpdeskslist {
			no := strconv.Itoa(i + 1)
			teamurl := strings.ReplaceAll(helpdesk, " ", "+")
			reply += no + ". Regional " + helpdesk + "\n" + "wa.me/" + Profile.Phonenumber + "?text=bantuan+operator+" + teamurl + "\n"
		}
		return
	}
	//suruh pilih scope dari bantuan team
	scope, scopelist, err := GetScopeFromTeam(Pesan, namateam, db)
	if err != nil {
		return err.Error()
	}
	//pilih scope jika belum
	if scope == "" {
		reply = "Terima kasih.\nSekarang, mohon pilih provinsi asal Bapak/Ibu dari daftar berikut:\n" // " + namateam + " :\n"
		for i, scope := range scopelist {
			no := strconv.Itoa(i + 1)
			scurl := strings.ReplaceAll(scope, " ", "+")
			reply += no + ". " + scope + "\n" + "wa.me/" + Profile.Phonenumber + "?text=bantuan+operator+" + namateam + "+" + scurl + "\n"
		}
		return
	}
	//menuliskan pertanyaan bantuan
	user = model.Laporan{
		Scope: scope,
		Team:  namateam,
		Nama:  Pesan.Alias_name,
		Phone: Pesan.Phone_number,
	}
	_, err = atdb.InsertOneDoc(db, "helpdeskuser", user)
	if err != nil {
		return err.Error()
	}
	reply = "Silakan ketik pertanyaan atau masalah yang ingin Bapak/Ibu " + Pesan.Alias_name + " sampaikan. Kami siap membantu Anda" // + " mengetik pertanyaan atau bantuan yang ingin dijawab oleh operator: "

	return
}

// handling key word
func FeedbackHelpdesk(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	msgs := strings.Split(Pesan.Message, "|")
	id := msgs[0]
	// Mengonversi id string ke primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		reply = "Invalid ID format: " + err.Error()
		return
	}
	helpdeskuser, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"_id": objectID, "phone": Pesan.Phone_number})
	if err != nil {
		reply = err.Error()
		return
	}
	strrate := strings.Split(msgs[1], ":")[1]
	rate := strings.TrimSpace(strrate)
	rt, err := strconv.Atoi(rate)
	if err != nil {
		reply = err.Error()
		return
	}
	helpdeskuser.RateLayanan = rt
	_, err = atdb.ReplaceOneDoc(db, "helpdeskuser", bson.M{"_id": objectID}, helpdeskuser)
	if err != nil {
		reply = err.Error()
		return
	}

	reply = "Terima kasih banyak atas waktu Bapak/Ibu untuk memberikan penilaian terhadap pelayanan Admin " + helpdeskuser.User.Name + "\n\nApresiasi Bapak/Ibu sangat berarti bagi kami untuk terus memberikan yang terbaik.."

	msgstr := "*Feedback Diterima*\n*" + helpdeskuser.Nama + "*\n*" + helpdeskuser.Phone + "*\nMemberikan rating " + rate + " bintang"
	dt := &itmodel.TextMessage{
		To:       helpdeskuser.User.PhoneNumber,
		IsGroup:  false,
		Messages: msgstr,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)

	return
}

// handling non key word
func PenugasanOperator(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string, err error) {
	//check apakah tiket dari user sudah di tutup atau belum
	user, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"phone": Pesan.Phone_number})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//check apakah dia operator yang belum tutup tiketnya
			user, err = atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"terlayani": bson.M{"$exists": false}, "user.phonenumber": Pesan.Phone_number})
			if err != nil {
				if err == mongo.ErrNoDocuments {
					err = nil
					reply = ""
					return
				}
				err = errors.New("galat di collection helpdeskuser operator: " + err.Error())
				return
			}
			//jika ada tiket yang statusnya belum closed
			reply = "*Permintaan bantuan dari Pengguna " + user.Nama + " (" + user.Phone + ")*\n\nMohon dapat segera menghubungi beliau melalui WhatsApp di nomor wa.me/" + user.Phone + " untuk memberikan solusi terkait masalah yang sedang dialami:\n\n" + user.Masalah
			reply += "\n\nSetelah masalah teratasi, dimohon untuk menginputkan solusi yang telah diberikan ke dalam sistem melalui tautan berikut:\nwa.me/" + Profile.Phonenumber + "?text=" + user.ID.Hex() + "|+solusi+dari+operator+helpdesk+:+"
			return

		}
		err = errors.New("galat di collection helpdeskuser user: " + err.Error())
		return
	}
	if !user.Terlayani {
		user.Masalah += "\n" + Pesan.Message
		if user.User.Name == "" || user.User.PhoneNumber == "" {
			var op model.Userdomyikado
			op, err = GetOperatorFromScopeandTeam(user.Scope, user.Team, db)
			if err != nil {
				return
			}
			user.User = op
		}
		_, err = atdb.ReplaceOneDoc(db, "helpdeskuser", bson.M{"_id": user.ID}, user)
		if err != nil {
			return
		}

		msgstr := "*Permintaan bantuan dari Pengguna " + user.Nama + " (" + user.Phone + ")*\n\nMohon dapat segera menghubungi beliau melalui WhatsApp di nomor wa.me/" + user.Phone + " untuk memberikan solusi terkait masalah yang sedang dialami:\n\n" + user.Masalah
		msgstr += "\n\nSetelah masalah teratasi, dimohon untuk menginputkan solusi yang telah diberikan ke dalam sistem melalui tautan berikut:\nwa.me/" + Profile.Phonenumber + "?text=" + user.ID.Hex() + "|+solusi+dari+operator+helpdesk+:+"
		dt := &itmodel.TextMessage{
			To:       user.User.PhoneNumber,
			IsGroup:  false,
			Messages: msgstr,
		}
		go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)

		reply = "Segera, Bapak/Ibu akan dihubungkan dengan salah satu Admin kami, *" + user.User.Name + "*.\n\n Mohon tunggu sebentar, kami akan menghubungi Anda melalui WhatsApp di nomor wa.me/" + user.User.PhoneNumber + "\nTerima kasih atas kesabaran Bapak/Ibu"

	}
	return

}
