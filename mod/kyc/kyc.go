package kyc

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CekSelfiePulang(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	if Pesan.Filedata == "" {
		return "Kirim pap nya dulu dong kak.. " + Pesan.Alias_name
	}
	dt := FaceDetect{
		IDUser:    Pesan.Phone_number,
		Base64Str: Pesan.Filedata,
	}
	filter := bson.M{"_id": atdb.TodayFilter(), "phonenumber": Pesan.Phone_number} //, "ismasuk": false}
	pstoday, err := atdb.GetOneDoc[PresensiLokasi](db, "presensi", filter)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf kakak belum cekin share live location hari ini " + err.Error()
	}
	conf, err := atdb.GetOneDoc[Config](db, "config", bson.M{"phonenumber": "62895601060000"})
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan dalam pengambilan config di database " + err.Error()
	}
	statuscode, faceinfo, err := atapi.PostStructWithToken[FaceInfo]("secret", conf.LeaflySecret, dt, conf.LeaflyURL)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan pemanggilan API leafly " + err.Error()
	}
	if statuscode != http.StatusOK {
		if statuscode == http.StatusFailedDependency {
			return "Wah kak " + Pesan.Alias_name + " mohon maaf, jangan kaku gitu dong. Ekspresi wajahnya ga boleh sama dengan selfie sebelumnya ya kak. Senyumnya yang lebar, giginya dilihatin, matanya pelototin, hidungnya keatasin.\n\n" + faceinfo.Error
		} else if statuscode == http.StatusMultipleChoices {
			return "IM$G#M$Gui76557u|||" + faceinfo.FileHash + "|||" + faceinfo.Error
		} else {
			return "Wah kak " + Pesan.Alias_name + " mohon maaf:\n" + faceinfo.Error + "\nCode: " + strconv.Itoa(statuscode)
		}

	}
	pselfie := PresensiSelfie{
		CekInLokasi: pstoday,
		IsMasuk:     false,
		IDUser:      faceinfo.PhoneNumber,
		Commit:      faceinfo.Commit,
		Filehash:    faceinfo.FileHash,
		Remaining:   faceinfo.Remaining,
	}
	_, err = atdb.InsertOneDoc(db, "selfie", pselfie)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan input ke database " + err.Error()
	}
	filter = bson.M{"_id": atdb.TodayFilter(), "cekinlokasi.phonenumber": Pesan.Phone_number, "ismasuk": true}
	selfiemasuk, err := atdb.GetOneLatestDoc[PresensiSelfie](db, "selfie", filter)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf kakak belum selfie masuk. " + err.Error()
	}
	// Ekstrak timestamp dari ObjectID
	objectIDTimestamp := selfiemasuk.ID.Timestamp()
	// Dapatkan waktu saat ini
	currentTime := time.Now()
	// Hitung selisih waktu dalam detik
	diff := currentTime.Sub(objectIDTimestamp) //.Seconds()
	// Konversi selisih waktu ke jam, menit, dan detik
	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60
	KetJam := fmt.Sprintf("%d jam, %d menit, %d detik", hours, minutes, seconds)

	skor := diff.Seconds() / 18000 //selisih waktu dibagi 8 jam
	skorValue := fmt.Sprintf("%f", skor)
	//post ke backedn domyikado
	datapresensi := PresensiDomyikado{
		ID:          selfiemasuk.ID,
		PhoneNumber: Pesan.Phone_number,
		Skor:        skor,
		KetJam:      KetJam,
		LamaDetik:   diff.Seconds(),
		Lokasi:      pstoday.Lokasi.Nama,
	}
	statuscode, httpresp, err := atapi.PostStructWithToken[itmodel.Response]("secret", conf.DomyikadoSecret, datapresensi, conf.DomyikadoPresensiURL)
	if err != nil {
		return "Akses ke endpoint domyikado gagal: " + err.Error()
	}
	if statuscode != http.StatusOK {
		return "Salah posting endpoint domyikado: " + httpresp.Response + "\ninfo\n" + httpresp.Info
	}
	return "Hai kak, " + Pesan.Alias_name + "\nBerhasil Presensi Pulang di lokasi:" + pstoday.Lokasi.Nama + "\nHadir selama: " + KetJam + "\n*Skor: " + skorValue + "*"

}

func CekKTP(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	if Pesan.Filedata == "" {
		return "Kirim pap nya dulu dong kak.. " + Pesan.Alias_name
	}
	dt := KTPProps{
		IDUser:    Pesan.Phone_number,
		Base64Str: Pesan.Filedata,
	}
	conf, err := atdb.GetOneDoc[Config](db, "config", bson.M{"phonenumber": Profile.Phonenumber})
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan dalam pengambilan config di database " + err.Error()
	}
	statuscode, faceinfo, err := atapi.PostStructWithToken[FaceInfo]("secret", conf.LeaflySecret, dt, conf.LeaflyURLKTP)
	if err != nil {
		return "Wah kak " + strconv.Itoa(statuscode) + " " + Pesan.Alias_name + " mohon maaf ada kesalahan pemanggilan API leafly :" + err.Error()
	}
	if statuscode != http.StatusOK {
		if statuscode == http.StatusFailedDependency {
			return "Wah kak " + Pesan.Alias_name + " mohon maaf, jangan kaku gitu dong. Ekspresi wajahnya ga boleh sama dengan selfie sebelumnya ya kak. Senyumnya yang lebar, giginya dilihatin, matanya pelototin, hidungnya keatasin.\n\n" + faceinfo.Error
		} else if statuscode == http.StatusMultipleChoices {
			return "IM$G#M$Gui76557u|||" + faceinfo.FileHash + "|||" + faceinfo.Error
		} else {
			return "Wah kak " + Pesan.Alias_name + " mohon maaf:\n" + faceinfo.Error + "\nCode: " + strconv.Itoa(statuscode)
		}

	}
	_, err = atdb.InsertOneDoc(db, "ktp", faceinfo)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan input ke database " + err.Error()
	}
	return "Hai kak, " + Pesan.Alias_name + ". KTP nya di cek ya dah valid.\n*Uuy testing*"

}

func PresensiMasuk(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	if !Pesan.LiveLoc {
		return "Minimal share live location dulu lah kak " + Pesan.Alias_name
	}
	longitude := fmt.Sprintf("%f", Pesan.Longitude)
	latitude := fmt.Sprintf("%f", Pesan.Latitude)
	lokasiuser, err := GetLokasi(db, Pesan.Longitude, Pesan.Latitude)
	if err != nil {
		return "Mohon maaf kak, kakak " + Pesan.Alias_name + " belum berada di lokasi presensi, silahkan menuju lokasi presensi dahulu baru cekin masuk."
	}
	if lokasiuser.Nama == "" {
		return "Nama nya kosong kak " + Pesan.Alias_name
	}
	dtuser := &PresensiLokasi{
		PhoneNumber: Pesan.Phone_number,
		Lokasi:      lokasiuser,
		IsMasuk:     true,
		CreatedAt:   time.Now(),
	}
	_, err = atdb.InsertOneDoc(db, "presensi", dtuser)
	if err != nil {
		return "Gagal insert ke database kak " + Pesan.Alias_name
	}

	return "Hai.. hai.. kakak atas nama:\n" + Pesan.Alias_name + "\nLongitude: " + longitude + "\nLatitude: " + latitude + "\nLokasi:" + lokasiuser.Nama + "\nsilahkan dilanjutkan dengan selfie di lokasi ya maximal 5 menit setelah share live location, jangan lupa ditambah keyword\n*myika selfie presensi masuk*"
}

func PresensiPulang(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	if !Pesan.LiveLoc {
		return "Minimal share live location dulu lah kak " + Pesan.Alias_name
	}
	longitude := fmt.Sprintf("%f", Pesan.Longitude)
	latitude := fmt.Sprintf("%f", Pesan.Latitude)
	lokasiuser, err := GetLokasi(db, Pesan.Longitude, Pesan.Latitude)
	if err != nil {
		return "Mohon maaf kak " + Pesan.Alias_name + ", kakak belum berada di lokasi presensi, silahkan menuju lokasi presensi dahulu baru cekin pulang."
	}
	if lokasiuser.Nama == "" {
		return "Nama nya kosong kak " + Pesan.Alias_name
	}
	dtuser := &PresensiLokasi{
		PhoneNumber: Pesan.Phone_number,
		Lokasi:      lokasiuser,
		IsMasuk:     false,
		CreatedAt:   time.Now(),
	}
	filter := bson.M{"_id": atdb.TodayFilter(), "cekinlokasi.phonenumber": Pesan.Phone_number, "ismasuk": true}
	docselfie, err := atdb.GetOneLatestDoc[PresensiSelfie](db, "selfie", filter)
	if err != nil {
		return "Kakak " + Pesan.Alias_name + " belum selfie masuk ini " + err.Error()
	}
	if docselfie.CekInLokasi.Lokasi.ID != lokasiuser.ID {
		return "Lokasi pulang nya harus sama dengan lokasi masuknya kak " + Pesan.Alias_name + ".\nLokasi : " + lokasiuser.Nama
	}
	_, err = atdb.InsertOneDoc(db, "presensi", dtuser)
	if err != nil {
		return "Gagal insert ke database kak " + Pesan.Alias_name
	}

	return "Hai.. hai.. kakak atas nama:\n" + Pesan.Alias_name + "\nLongitude: " + longitude + "\nLatitude: " + latitude + "\nLokasi:" + lokasiuser.Nama + "\nsilahkan dilanjutkan dengan selfie di lokasi ya maximal 5 menit setelah share live location, jangan lupa ditambah keyword\n*myika selfie presensi pulang*"
}

func GetLokasi(mongoconn *mongo.Database, long float64, lat float64) (lokasi Lokasi, err error) {
	filter := bson.M{
		"batas": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{long, lat},
				},
			},
		},
	}

	lokasi, err = atdb.GetOneDoc[Lokasi](mongoconn, "lokasi", filter)
	if err != nil {
		return
	}
	return
}
