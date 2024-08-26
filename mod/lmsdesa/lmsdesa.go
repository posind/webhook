package lmsdesa

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ArsipGambar(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	if Pesan.Filedata == "" {
		return "Kirim nya dengan gambar nya dulu dong kak.. " + Pesan.Alias_name
	}
	dt := FaceDetect{
		IDUser:    Pesan.Phone_number,
		IDFile:    Pesan.Filename,
		Base64Str: Pesan.Filedata,
	}
	conf, err := atdb.GetOneDoc[Config](db, "config", bson.M{"phonenumber": "62895601060000"})
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan dalam pengambilan config di database " + err.Error()
	}
	statuscode, faceinfo, err := atapi.PostStructWithToken[FaceInfo]("secret", conf.LeaflySecret, dt, conf.LeaflyURLLMSDesaGambar)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan pemanggilan API leafly: " + err.Error()
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
	return "Hai kak, " + Pesan.Alias_name + "\nBerhasil simpan gambar dengan hash:" + faceinfo.FileHash + "\nNomor Commit: " + faceinfo.Commit + "\n*Remaining: " + strconv.Itoa(faceinfo.Remaining) + "*"

}

func ArsipFile(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	if Pesan.Filedata == "" {
		return "Kirim nya dengan gambar nya dulu dong kak.. " + Pesan.Alias_name
	}
	dt := FaceDetect{
		IDUser:    Pesan.Phone_number,
		IDFile:    Pesan.Filename,
		Base64Str: Pesan.Filedata,
	}
	conf, err := atdb.GetOneDoc[Config](db, "config", bson.M{"phonenumber": "62895601060000"})
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan dalam pengambilan config di database " + err.Error()
	}
	statuscode, faceinfo, err := atapi.PostStructWithToken[FaceInfo]("secret", conf.LeaflySecret, dt, conf.LeaflyURLLMSDesaFile)
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan pemanggilan API leafly: " + err.Error()
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
	tagmdimg := "![image](" + filepath.Base(faceinfo.FileHash) + ")"
	return "Hai kak, " + Pesan.Alias_name + "\nSilahkan tempelkan script berikut pada resume notulen meeting untuk menyisipkan gambar ini:\n\n*" + tagmdimg + "*\n\nNomor Commit: " + faceinfo.Commit + "\nRemaining: " + strconv.Itoa(faceinfo.Remaining)

}
