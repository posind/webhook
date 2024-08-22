package posint

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
				"Destination":      country,
				"Prohibited Items": bson.M{"$regex": keyword, "$options": "i"},
			}
		} else {
			filter = bson.M{"Destination": country}
		}
		reply, _, err = populateList(db, filter, keyword)
		reply = "💡" + reply
		if err != nil {
			if err.Error() == "zero results" {
				return " is allowed to send to " + country
			}
			jsonData, _ := bson.Marshal(filter)
			return "💡" + countryandkeyword + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
		}
		return
	}
	return
}

func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listprob, err := atdb.GetAllDoc[[]ItemProhibited](db, "prohibited_items_en", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listprob) == 0 {
		return " There is no prohibited items that found!", "", errors.New("zero results")
	}
	dest = listprob[0].Destination
	msg = " Here is a list of prohibited items from the country *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci: _" + keyword + "_\n"
	}
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.ProhibitedItems + "\n"
	}
	return msg, dest, nil
}

func GetCountryNameLike(db *mongo.Database, country string) (dest string, err error) {
	filter := bson.M{
		"Destination": bson.M{"$regex": country, "$options": "i"},
	}
	itemprohb, err := atdb.GetOneDoc[ItemProhibited](db, "prohibited_items_en", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(itemprohb.Destination, "\u00A0", " ")
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

// ExtractKeywords Berungsi untuk menghilangkan semua kata kecuali keyword yang diinginkan
func ExtractKeywords(message string, commonWordsAdd []string) string {
	// Daftar kata umum yang mungkin ingin dihilangkan
	commonWords := []string{"list", "en", "mymy"}

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
