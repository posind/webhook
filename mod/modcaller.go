package mod

import (
	"github.com/gocroot/mod/countrylist"
	"github.com/gocroot/mod/daftar"
	"github.com/gocroot/mod/idgrup"
	"github.com/gocroot/mod/kyc"
	"github.com/gocroot/mod/lmsdesa"
	"github.com/gocroot/mod/posint"
	posintid "github.com/gocroot/mod/posintind"
	"github.com/gocroot/mod/presensi"
	"github.com/gocroot/mod/siakad"
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
	case "login-siakad":
		reply = siakad.LoginSiakad(Pesan, db)
	case "prohibited-items":
		reply = posint.GetProhibitedItems(Pesan, db)
	case "prohibited-items-id":
		reply = posintid.GetProhibitedItems(Pesan, db)
	case "countrylist":
		reply = countrylist.CountryList(Pesan)
	default:
		reply = "Modul tidak ditemukan"
	}
	return
}
