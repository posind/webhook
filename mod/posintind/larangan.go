package posintind

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/kimseok"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	country, _, _, err := kimseok.GetCountryFromMessage(Pesan.Message, db)
	var filter bson.M
	var keyword string
	if err != nil {
		countryandkeyword := ExtractKeywords(Pesan.Message, []string{})
		words := strings.Split(countryandkeyword, " ")
		var key []string
		// Iterate through the slice, popping elements from the end
		for len(words) > 0 {
			// Join remaining elements back into a string
			remainingMessage := strings.Join(words, " ")
			country, err = GetCountryNameLike(db, remainingMessage)
			if err == nil {
				break
			}
			// Get the last element
			lastWord := words[len(words)-1]
			key = append(key, lastWord)
			// Remove the last element
			words = words[:len(words)-1]
		}
		if len(key) > 0 {
			keyword = strings.Join(key, " ")
			filter = bson.M{
				"Destinasi":        country,
				"Barang Terlarang": bson.M{"$regex": keyword, "$options": "i"},
			}
		} else {
			filter = bson.M{"Destinasi": country}
		}
		reply, _, err = populateList(db, filter, keyword)
		reply = "💡" + reply
		if err != nil {
			if err.Error() == "zero results" {
				return " diperbolehkan untuk dikirim ke negara " + country
			}
			jsonData, _ := bson.Marshal(filter)
			return "💡" + countryandkeyword + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
		}
		return
	}
	if country == "" {
		return "Nama negara tidak ada kak di database kita"
	}
	keyword = ExtractKeywords(Pesan.Message, []string{country})
	if keyword != "" {
		filter = bson.M{
			"Destinasi":        country,
			"Barang Terlarang": bson.M{"$regex": keyword, "$options": "i"},
		}
	} else {
		filter = bson.M{"Destinasi": country}
	}
	reply, _, err = populateList(db, filter, keyword)
	reply = "📚" + reply
	if err != nil {
		if err.Error() == "zero results" {
			return "📚" + keyword + " diperbolehkan untuk dikirim ke negara " + country
		}
		jsonData, _ := bson.Marshal(filter)
		return "📚 " + keyword + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
	}
	return
}

func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listprob, err := atdb.GetAllDoc[[]Item](db, "prohibited_items_id", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listprob) == 0 {
		return " Tidak ada barang terlarang yang ditemukan", "", errors.New("zero results")
	}
	dest = listprob[0].Destinasi
	msg = " Ini dia list barang terlarang dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci: _" + keyword + "_\n"
	}
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.BarangTerlarang + "\n"
	}
	return msg, dest, nil
}

func GetCountryNameLike(db *mongo.Database, country string) (dest string, err error) {
	filter := bson.M{
		"Destinasi": bson.M{"$regex": country, "$options": "i"},
	}
	itemprohb, err := atdb.GetOneDoc[Item](db, "prohibited_items_id", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(itemprohb.Destinasi, "\u00A0", " ")
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
	countries, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destinasi", "prohibited_items_id")
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
	commonWords := []string{"list", "id", "mymy"}

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
