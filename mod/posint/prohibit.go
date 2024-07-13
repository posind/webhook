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
	var filter bson.M
	var dest string
	if err != nil {
		countryandkeyword := ExtractKeywords(Pesan.Message, []string{})
		keywords := strings.Split(countryandkeyword, " ")
		country := keywords[0]
		filter = bson.M{
			"Destination": bson.M{"$regex": country, "$options": "i"},
		}
		reply, dest, err = populateList(db, filter)
		if err != nil {
			return dest + " " + err.Error()
		}
		return
	}
	if country == "" {
		return "Nama negara tidak ada kak di database kita"
	}
	keyword := ExtractKeywords(Pesan.Message, []string{country})
	if keyword != "" {
		filter = bson.M{
			"Destination":      country,
			"Prohibited Items": bson.M{"$regex": keyword, "$options": "i"},
		}
	} else {
		filter = bson.M{"Destination": country}
	}
	reply, dest, err = populateList(db, filter)
	if err != nil {
		return dest + " " + err.Error()
	}
	return

}

func populateList(db *mongo.Database, filter bson.M) (msg string, dest string, err error) {
	listprob, err := atdb.GetAllDoc[[]Item](db, "prohibited_items_en", filter)
	if err != nil {
		return "Terdapat kesalahan pada  GetAllDoc ", "", err
	}
	if len(listprob) == 0 {
		return "Tidak ada prohibited items yang ditemukan ", "", errors.New("zero results")
	}
	dest = listprob[0].Destination
	msg = "ini dia list prohibited item dari negara *" + dest + "*:\n"
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.ProhibitedItems + "\n"
	}
	return
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
