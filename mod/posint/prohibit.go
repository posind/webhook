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
	if err != nil {
		countryandkeyword := ExtractKeywords(Pesan.Message, []string{})
		keywords := strings.Split(countryandkeyword, " ")
		if len(keywords) == 0 {
			return "Nama negara tidak ada kak di database kita"
		} else if len(keywords) > 2 {
			country = keywords[0] + " " + keywords[1]

		} else if len(keywords) == 1 {
			country = keywords[0]
		}
		//query dulu nama country yang bener di db dari yang mirip regex
		filter = bson.M{
			"Destination": bson.M{"$regex": country, "$options": "i"},
		}
		var dest string
		_, dest, err = populateList(db, filter)
		//reply = "💡" + reply
		if err != nil {
			jsonData, _ := bson.Marshal(filter)
			return "💡" + countryandkeyword + "|" + country + " : " + err.Error() + string(jsonData)
		}
		keyword := ExtractKeywords(Pesan.Message, []string{dest})
		if keyword != "" {
			filter = bson.M{
				"Destination":      dest,
				"Prohibited Items": bson.M{"$regex": keyword, "$options": "i"},
			}
		} else {
			filter = bson.M{"Destination": dest}
		}
		reply, _, err = populateList(db, filter)
		reply = "💡" + reply
		if err != nil {
			jsonData, _ := bson.Marshal(filter)
			return "💡" + countryandkeyword + "|" + country + " : " + err.Error() + string(jsonData)
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
	reply, _, err = populateList(db, filter)
	reply = "📚" + reply
	if err != nil {
		jsonData, _ := bson.Marshal(filter)
		return "📚" + keyword + "|" + country + " : " + err.Error() + string(jsonData)
	}
	return

}

func populateList(db *mongo.Database, filter bson.M) (msg, dest string, err error) {
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

func GetCountryNameLike(db *mongo.Database, country string) (dest string, err error) {
	filter := bson.M{
		"Destination": bson.M{"$regex": country, "$options": "i"},
	}
	itemprohb, err := atdb.GetOneDoc[Item](db, "prohibited_items_en", filter)
	if err != nil {
		return
	}
	dest = itemprohb.Destination
	return
}

func GetCountryFromMessage(message string, db *mongo.Database) (country string, err error) {
	// Ubah pesan menjadi huruf kecil
	lowerMessage := strings.ToLower(message)
	// Mengganti non-breaking space dengan spasi biasa
	lowerMessage = strings.ReplaceAll(lowerMessage, "\u00A0", " ")
	// Hapus spasi berlebih
	lowerMessage = strings.TrimSpace(lowerMessage)
	lowerMessage = regexp.MustCompile(`\s+`).ReplaceAllString(lowerMessage, " ")
	// Mendapatkan nama negara
	countries, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destination", "prohibited_items_en")
	if err != nil {
		return "", err
	}
	var strcountry string
	// Iterasi melalui daftar negara
	for _, country := range countries {
		lowerCountry := strings.ToLower(strings.TrimSpace(country.(string)))
		// Mengganti non-breaking space dengan spasi biasa
		lowerCountry = strings.ReplaceAll(lowerCountry, "\u00A0", " ")
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
	commonWords := []string{"list", "prohibited", "items", "item", "mymy"}

	// Gabungkan commonWords dengan commonWordsAdd
	commonWords = append(commonWords, commonWordsAdd...)

	// Ubah pesan menjadi huruf kecil
	message = strings.ToLower(message)

	// Ganti non-breaking space dengan spasi biasa
	message = strings.ReplaceAll(message, "\u00A0", " ")

	// Hapus kata-kata umum dari pesan
	for _, word := range commonWords {
		word = strings.ToLower(strings.ReplaceAll(word, "\u00A0", " "))
		message = strings.ReplaceAll(message, word, "")
	}

	// Hapus spasi berlebih
	message = strings.TrimSpace(message)
	message = regexp.MustCompile(`\s+`).ReplaceAllString(message, " ")

	return message
}
