package daftar

import (
	"net/http"
	"regexp"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func DaftarDomyikado(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	// Define a regex pattern for email addresses
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

	// Find the email address in the string
	email := re.FindString(Pesan.Message)
	if email == "user@email.com" {
		return "Emailnya di ubah dulu dong kak, jadi emailnya kak " + Pesan.Alias_name
	} else if email == "" {
		return "Emailnya di sertakan dulu dong kak " + Pesan.Alias_name + " di akhir pesan nya"
	}

	conf, err := atdb.GetOneDoc[Config](db, "config", bson.M{"phonenumber": "62895601060000"})
	if err != nil {
		return "Wah kak " + Pesan.Alias_name + " mohon maaf ada kesalahan dalam pengambilan config di database " + err.Error()
	}
	//post ke backedn domyikado
	calonuser := Userdomyikado{
		PhoneNumber: Pesan.Phone_number,
		Name:        Pesan.Alias_name,
		Email:       email,
	}

	statuscode, httpresp, err := atapi.PostStructWithToken[itmodel.Response]("secret", conf.DomyikadoSecret, calonuser, conf.DomyikadoUserURL)
	if err != nil {
		return "Akses ke endpoint domyikado gagal: " + err.Error()
	}
	if statuscode != http.StatusOK {
		return "Salah posting endpoint domyikado: " + httpresp.Response + "\ninfo\n" + httpresp.Info
	}
	return "Hai kak, " + Pesan.Alias_name + "\nBerhasil didaftarkan dengan email:" + email

}
