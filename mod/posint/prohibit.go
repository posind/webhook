package posint

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	country, err := GetCountryFromMessage(Pesan.Message, db)
	if err != nil {
		countryandkeyword := ExtractKeywords(Pesan.Message, []string{})
		return countryandkeyword + "|" + err.Error()
	}
	if country == "" {
		return "Nama negara tidak ada kak di database kita"
		//countryandkeyword := ExtractKeywords(Pesan.Message, []string{})
		//countryname, err := atdb.GetOneDoc[Item](db, "prohibited_items_en", bson.M{"Destination": bson.M{"$regex": countryandkeyword, "$options": "i"}})
	}

	keyword := ExtractKeywords(Pesan.Message, []string{country})
	var filter bson.M
	if keyword != "" {
		filter = bson.M{
			"Destination":      country,
			"Prohibited Items": bson.M{"$regex": keyword, "$options": "i"},
		}
	} else {
		filter = bson.M{"Destination": country}
	}
	listprob, err := atdb.GetAllDoc[[]Item](db, "prohibited_items_en", filter)
	if err != nil {
		return "Terdapat kesalahan pada  GetAllDoc " + err.Error()
	}
	if len(listprob) == 0 {
		return "Tidak ada prohibited items yang ditemukan untuk negara " + country + " dengan keyword " + keyword
	}
	msg := "ini dia list prohibited item dari negara yang kakak minta:\n"
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.ProhibitedItems + "\n"
	}
	return msg

}

func GetCountryFromMessage(message string, db *mongo.Database) (country string, err error) {
	// Ubah pesan menjadi huruf kecil
	lowerMessage := strings.ToLower(message)
	// Mendapatkan nama negara
	countries, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destination", "prohibited_items_en")
	if err != nil {
		return "", err
	}
	var strcountry string
	// Iterasi melalui daftar negara
	for _, country := range countries {
		lowerCountry := strings.ToLower(strings.TrimSpace(country.(string)))
		strcountry += lowerCountry + ","
		if strings.Contains(lowerMessage, lowerCountry) {
			return country.(string), nil
		}
	}
	return "", errors.New("tidak ditemukan nama negara di pesan berikut:" + lowerMessage + "|" + strcountry)
}

// Fungsi untuk menghilangkan semua kata kecuali keyword yang diinginkan
func ExtractKeywords(message string, commonWordsAdd []string) string {
	// Daftar kata umum yang mungkin ingin dihilangkan
	commonWords := []string{"list", "prohibited", "items", "myika"}

	// Gabungkan commonWords dengan commonWordsAdd
	commonWords = append(commonWords, commonWordsAdd...)

	// Ubah pesan menjadi huruf kecil
	message = strings.ToLower(message)

	// Hapus kata-kata umum dari pesan
	for _, word := range commonWords {
		message = strings.ReplaceAll(message, strings.ToLower(word), "")
	}

	// Hapus spasi berlebih
	message = strings.TrimSpace(message)
	message = regexp.MustCompile(`\s+`).ReplaceAllString(message, " ")

	return message
}
