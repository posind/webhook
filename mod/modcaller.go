package mod

import (
	"strings"

	"github.com/gocroot/helper/kimseok"
	"github.com/gocroot/mod/daftar"
	// "github.com/gocroot/mod/helpdesk"
	"github.com/gocroot/mod/idgrup"
	"github.com/gocroot/mod/kyc"
	// "github.com/gocroot/mod/listcountry"
	// "github.com/gocroot/mod/listnegara"
	"github.com/gocroot/mod/lmsdesa"
	"github.com/gocroot/mod/maxweight"
	"github.com/gocroot/mod/posint"
	"github.com/gocroot/mod/presensi"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/mongo"
)

func Caller(Profile itmodel.Profile, Modulename string, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	switch Modulename {
	case "idgrup":
		reply = idgrup.IDGroup(Pesan)
	case "presensi-masuk":
		reply = presensi.PresensiMasuk(Pesan, db)
	case "presensi-pulang":
		reply = presensi.PresensiPulang(Pesan, db)
	case "upload-lmsdesa-file":
		reply = lmsdesa.ArsipFile(Pesan, db)
	case "upload-lmsdesa-gambar":
		reply = lmsdesa.ArsipGambar(Pesan, db)
	case "cek-ktp":
		reply = kyc.CekKTP(Profile, Pesan, db)
	case "selfie-masuk":
		reply = presensi.CekSelfieMasuk(Profile, Pesan, db)
	case "selfie-pulang":
		reply = presensi.CekSelfiePulang(Pesan, db)
	case "domyikado-user":
		reply = daftar.DaftarDomyikado(Pesan, db)

	case "prohibited-items":
		reply = posint.GetProhibitedItems(Pesan, db)
	case "max-weight":
		reply = maxweight.GetMaxWeight(Pesan, db)

	// case "listcountry":
	// 	reply = listcountry.ListCountry(Pesan)
	// case "listnegara":
	// 	reply = listnegara.ListNegara(Pesan)

	// case "feedbackhelpdesk":
	// 	reply = helpdesk.FeedbackHelpdesk(Profile, Pesan, db)
	// case "endhelpdesk":
	// 	reply = helpdesk.EndHelpdesk(Profile, Pesan, db)
	// case "helpdesk":
	// 	reply = helpdesk.HelpdeskPDLMS(Profile, Pesan, db)
	// case "helpdeskpusat":
	// 	reply = helpdesk.HelpdeskPusat(Profile, Pesan, db)
	// case "adminopenusertiket":
	// 	reply = helpdesk.AdminOpenSessionCurrentUserTiket(Profile, Pesan, db)

	default:
		// Fallback to QueriesDataRegexpALL if no case matches
		dt, err := kimseok.QueriesDataRegexpALL(db, Pesan.Message)
		if err != nil {
			reply = "Maaf, terjadi kesalahan saat mencari data."
		} else {
			reply = strings.TrimSpace(dt.Answer) // Use the result from QueriesDataRegexpALL
		}
	}
	return
}
