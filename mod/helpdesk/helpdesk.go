package helpdesk

import (
	"fmt"
	"strings"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/hub"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/helper/menu"
	"github.com/gocroot/helper/phone"
	"github.com/gocroot/helper/tiket"
	"github.com/gocroot/helper/waktu"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// helpdesk sudah terintegrasi dengan lms pamong desa backend
func HelpdeskPDLMS(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	//check apakah tiketnya udah tutup atau belum
	isclosed, stiket, err := tiket.IsTicketClosed("userphone", Pesan.Phone_number, db)
	if err != nil {
		return "IsTicketClosed: " + err.Error()
	}
	if !isclosed { //ada yang belum closed, lanjutkan sesi hub
		//pesan ke user
		reply = GetPrefillMessage("userbantuanadmin", db) //pesan ke user
		reply = fmt.Sprintf(reply, stiket.AdminName)
		hub.CheckHubSession(stiket.UserPhone, stiket.UserName, stiket.AdminPhone, stiket.AdminName, db)
		//inject menu session untuk menutup tiket
		mn := menu.MenuList{
			No:      0,
			Keyword: stiket.ID.Hex() + "|tutuph3lpdeskt1kcet",
			Konten:  "Akhiri percakapan dan tutup sesi bantuan saat ini",
		}
		err = menu.InjectSessionMenu([]menu.MenuList{mn}, stiket.UserPhone, db)
		if err != nil {
			return err.Error()
		}
		err = menu.InjectSessionMenu([]menu.MenuList{mn}, stiket.AdminPhone, db)
		if err != nil {
			return err.Error()
		}
		return
	}
	//jika tiket sudah clear
	statuscode, res, err := atapi.GetStructWithToken[lms.ResponseAPIPD]("token", config.APITOKENPD, config.APIGETPDLMS+Pesan.Phone_number)
	if statuscode != 200 { //404 jika user not found
		msg := "Mohon maaf Bapak/Ibu, nomor anda *belum terdaftar* pada sistem kami.\n" + UserNotFound(Profile, Pesan, db)
		return msg
	}
	if err != nil {
		return err.Error()
	}
	if len(res.Data.ContactAdminProvince) == 0 { //kalo kosong data kontak admin provinsinya maka arahkan ke tim 16 tapi sesuikan dengan provinsinya
		msg := "Mohon maaf Bapak/Ibu " + res.Data.Fullname + " dari desa " + res.Data.Village + ", helpdesk pamongdesa anda.\n" + AdminNotFoundWithProvinsi(Profile, Pesan, res.Data.Province, db)
		return msg
	}
	//jika arraynya ada adminnya maka lanjut ke start session hub
	helpdeskno := res.Data.ContactAdminProvince[0].Phone
	helpdeskname := res.Data.ContactAdminProvince[0].Fullname
	if helpdeskname == "" || helpdeskno == "" {
		return "Nama atau nomor helpdesk tidak ditemukan"
	}
	//pesan ke admin
	msgstr := GetPrefillMessage("adminbantuanadmin", db) //pesan ke admin
	msgstr = fmt.Sprintf(msgstr, res.Data.Fullname, res.Data.Village, res.Data.District, res.Data.Regency)
	dt := &itmodel.TextMessage{
		To:       helpdeskno,
		IsGroup:  false,
		Messages: msgstr,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)
	//pesan ke user
	reply = GetPrefillMessage("userbantuanadmin", db) //pesan ke user
	reply = fmt.Sprintf(reply, helpdeskname)
	//insert ke database dan set hub session
	idtiket, err := tiket.InserNewTicket(Pesan.Phone_number, helpdeskname, helpdeskno, db)
	if err != nil {
		return err.Error()
	}
	hub.CheckHubSession(Pesan.Phone_number, res.Data.Fullname, helpdeskno, helpdeskname, db)
	//inject menu session untuk menutup tiket
	mn := menu.MenuList{
		No:      0,
		Keyword: idtiket.Hex() + "|tutuph3lpdeskt1kcet",
		Konten:  "Akhiri percakapan dan tutup sesi bantuan saat ini",
	}
	err = menu.InjectSessionMenu([]menu.MenuList{mn}, Pesan.Phone_number, db)
	if err != nil {
		return err.Error()
	}
	err = menu.InjectSessionMenu([]menu.MenuList{mn}, helpdeskno, db)
	if err != nil {
		return err.Error()
	}
	return

}

// Jika user tidak terdaftar maka akan mengeluarkan list operator pusat
func UserNotFound(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	//check apakah ada session, klo ga ada kasih reply menu
	Sesdoc, _, err := menu.CheckSession(Pesan.Phone_number, db)
	if err != nil {
		return err.Error()
	}

	msg, err := menu.GetMenuFromKeywordAndSetSession("adminpusat", Sesdoc, db)
	if err != nil {
		return err.Error()
	}
	return msg
}

// penugasan helpdeskpusat jika user belum terdaftar, ini limpahan dari pilihan func UserNotFound
func HelpdeskPusat(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	Pesan.Message = strings.ReplaceAll(Pesan.Message, "adminpusat", "")
	Pesan.Message = strings.TrimSpace(Pesan.Message)
	op, err := GetOperatorFromSection(Pesan.Message, db)
	if err != nil {
		return err.Error()
	}
	res := lms.GetDataFromAPI(Pesan.Phone_number)
	//pesan untuk admin
	msgstr := GetPrefillMessage("adminbantuanadmin", db)
	if res.Data.Fullname != "" {
		msgstr = fmt.Sprintf(msgstr, res.Data.Fullname, res.Data.Village, res.Data.District, res.Data.Regency)
	} else {
		msgstr = fmt.Sprintf(msgstr, phone.MaskPhoneNumber(Pesan.Phone_number)+" ~ "+Pesan.Alias_name, "belum", "terdaftar", "sistem")
	}
	dt := &itmodel.TextMessage{
		To:       op.PhoneNumber,
		IsGroup:  false,
		Messages: msgstr,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)
	//pesan untuk user
	reply = GetPrefillMessage("userbantuanadmin", db)
	reply = fmt.Sprintf(reply, op.Name)
	//insert ke database dan set hub session
	idtiket, err := tiket.InserNewTicket(Pesan.Phone_number, op.Name, op.PhoneNumber, db)
	if err != nil {
		return err.Error()
	}
	hub.CheckHubSession(Pesan.Phone_number, phone.MaskPhoneNumber(Pesan.Phone_number)+" ~ "+Pesan.Alias_name, op.PhoneNumber, op.Name, db)
	//inject menu session untuk menutup tiket
	mn := menu.MenuList{
		No:      0,
		Keyword: idtiket.Hex() + "|tutuph3lpdeskt1kcet",
		Konten:  "Akhiri percakapan dan tutup sesi bantuan saat ini",
	}
	err = menu.InjectSessionMenu([]menu.MenuList{mn}, Pesan.Phone_number, db)
	if err != nil {
		return err.Error()
	}
	err = menu.InjectSessionMenu([]menu.MenuList{mn}, op.PhoneNumber, db)
	if err != nil {
		return err.Error()
	}
	return

}

// Jika user terdaftar tapi belum ada operator provinsi maka akan mengeluarkan list operator pusat
func AdminNotFoundWithProvinsi(Profile itmodel.Profile, Pesan itmodel.IteungMessage, provinsi string, db *mongo.Database) (reply string) {
	//tambah lojik query ke provinsi
	sec, err := GetSectionFromProvinsiRegex(db, provinsi)
	if err != nil {
		return err.Error()
	}
	op, err := GetOperatorFromSection(sec, db)
	if err != nil {
		return err.Error()
	}
	res := lms.GetDataFromAPI(Pesan.Phone_number)
	msgstr := GetPrefillMessage("adminbantuanadmin", db) //pesan untuk admin
	msgstr = fmt.Sprintf(msgstr, res.Data.Fullname, res.Data.Village, res.Data.District, res.Data.Regency)
	dt := &itmodel.TextMessage{
		To:       op.PhoneNumber,
		IsGroup:  false,
		Messages: msgstr,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)
	reply = GetPrefillMessage("userbantuanadmin", db) //pesan untuk user
	reply = fmt.Sprintf(reply, op.Name)
	//insert ke database dan set hub session
	idtiket, err := tiket.InserNewTicket(Pesan.Phone_number, op.Name, op.PhoneNumber, db)
	if err != nil {
		return err.Error()
	}
	hub.CheckHubSession(Pesan.Phone_number, res.Data.Fullname, op.PhoneNumber, op.Name, db)
	//inject menu session untuk menutup tiket
	mn := menu.MenuList{
		No:      0,
		Keyword: idtiket.Hex() + "|tutuph3lpdeskt1kcet",
		Konten:  "Akhiri percakapan dan tutup sesi bantuan saat ini",
	}
	err = menu.InjectSessionMenu([]menu.MenuList{mn}, Pesan.Phone_number, db)
	if err != nil {
		return err.Error()
	}
	err = menu.InjectSessionMenu([]menu.MenuList{mn}, op.PhoneNumber, db)
	if err != nil {
		return err.Error()
	}
	return
}

// penutupan helpdesk dari pilihan menu objectid|tutuph3lpdeskt1kcet
func EndHelpdesk(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	msgs := strings.Split(Pesan.Message, "|")
	id := msgs[0]
	// Mengonversi id string ke primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		reply = "Invalid ID format: " + err.Error()
		return
	}
	helpdeskuser, err := atdb.GetOneLatestDoc[tiket.Bantuan](db, "tiket", bson.M{"_id": objectID})
	if err != nil {
		reply = err.Error()
		return
	}
	//helpdeskuser.Solusi = strings.Split(msgs[1], ":")[1]
	helpdeskuser.Terlayani = true
	helpdeskuser.CloseAt = waktu.Sekarang()
	_, err = atdb.ReplaceOneDoc(db, "tiket", bson.M{"_id": objectID}, helpdeskuser)
	if err != nil {
		reply = err.Error()
		return
	}
	//hapus hub
	atdb.DeleteOneDoc(db, "hub", bson.M{"userphone": helpdeskuser.UserPhone, "adminphone": helpdeskuser.AdminPhone})
	//hapus session menu
	atdb.DeleteOneDoc(db, "session", bson.M{"phonenumber": helpdeskuser.UserPhone})
	atdb.DeleteOneDoc(db, "session", bson.M{"phonenumber": helpdeskuser.AdminPhone})
	//prefill message admin dan user
	msgstradmin := GetPrefillMessage("admintutuphelpdesk", db) //pesan untuk admin
	if helpdeskuser.UserName != "" {
		msgstradmin = fmt.Sprintf(msgstradmin, helpdeskuser.UserName, helpdeskuser.Desa)
	} else {
		msgstradmin = fmt.Sprintf(msgstradmin, phone.MaskPhoneNumber(helpdeskuser.UserPhone), "Belum Terdaftar Sistem")
	}

	msgstruser := GetPrefillMessage("usertutuphelpdesk", db) //pesan untuk user
	msgstruser = fmt.Sprintf(msgstruser, helpdeskuser.AdminName, helpdeskuser.UserName, helpdeskuser.ID.Hex())
	//pembagian yg dikirim dan reply
	var sendmsg, to string
	if Pesan.Phone_number == helpdeskuser.UserPhone {
		reply = msgstruser
		sendmsg = msgstradmin
		to = helpdeskuser.AdminPhone
	} else {
		reply = msgstradmin
		sendmsg = msgstruser
		to = helpdeskuser.UserPhone
	}
	dt := &itmodel.TextMessage{
		To:       to,
		IsGroup:  false,
		Messages: sendmsg,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)

	return
}

// admin terkoneksi dengan user tiket terakhir yang belum terlayani
func AdminOpenSessionCurrentUserTiket(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	//check apakah tiketnya udah tutup atau belum
	isclosed, stiket, err := tiket.IsTicketClosed("adminphone", Pesan.Phone_number, db)
	if err != nil {
		return "IsTicketClosed: " + err.Error()
	}
	if !isclosed { //ada yang belum closed, lanjutkan sesi hub
		//pesan ke admin
		reply = GetPrefillMessage("adminadasesitiket", db)
		reply = fmt.Sprintf(reply, stiket.UserName, stiket.Desa, stiket.Kec, stiket.KabKot)
		hub.CheckHubSession(stiket.UserPhone, stiket.UserName, stiket.AdminPhone, stiket.AdminName, db)
		//inject menu session untuk menutup tiket
		mn := menu.MenuList{
			No:      0,
			Keyword: stiket.ID.Hex() + "|tutuph3lpdeskt1kcet",
			Konten:  "Akhiri percakapan dan tutup sesi bantuan saat ini",
		}
		err = menu.InjectSessionMenu([]menu.MenuList{mn}, stiket.AdminPhone, db)
		if err != nil {
			return err.Error()
		}
		err = menu.InjectSessionMenu([]menu.MenuList{mn}, stiket.UserPhone, db)
		if err != nil {
			return err.Error()
		}
		return
	}
	//pesan ke admin
	reply = GetPrefillMessage("adminkosongsesitiket", db)
	reply = fmt.Sprintf(reply, stiket.AdminName)
	return
}
