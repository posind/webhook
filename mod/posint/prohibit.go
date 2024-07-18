package posint

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"github.com/xrash/smetrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	country, err := GetCountryFromMessage(Pesan.Message, db)
	var filter bson.M
	if err != nil {
		keywords := ExtractKeywords(Pesan.Message, []string{})
		words := strings.Split(strings.Join(keywords, " "), " ")
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
			keyword := strings.Join(key, " ")
			regexPattern := BuildFlexibleRegex(ExtractKeywords(keyword, []string{}))
			filter = bson.M{
				"Destination":      country,
				"Prohibited Items": bson.M{"$regex": regexPattern, "$options": "i"},
			}
		} else {
			filter = bson.M{"Destination": country}
		}
		reply, _, err = populateList(db, filter, strings.Join(key, " "))
		reply = "💡" + reply
		if err != nil {
			jsonData, _ := bson.Marshal(filter)
			return "💡" + strings.Join(keywords, " ") + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
		}
		return
	}
	if country == "" {
		return "Nama negara tidak ada kak di database kita"
	}
	keywords := ExtractKeywords(Pesan.Message, []string{country})
	if len(keywords) > 0 {
		regexPattern := BuildFlexibleRegex(keywords)
		filter = bson.M{
			"Destination":      country,
			"Prohibited Items": bson.M{"$regex": regexPattern, "$options": "i"},
		}
	} else {
		filter = bson.M{"Destination": country}
	}
	reply, _, err = populateList(db, filter, strings.Join(keywords, " "))
	reply = "📚" + reply
	if err != nil {
		jsonData, _ := bson.Marshal(filter)
		return "📚" + strings.Join(keywords, " ") + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
	}
	return
}

func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listprob, err := atdb.GetAllDoc[[]Item](db, "prohibited_items_en", filter)
	if err != nil {
		return "Terdapat kesalahan pada  GetAllDoc ", "", err
	}
	if len(listprob) == 0 {
		return "Tidak ada prohibited items yang ditemukan ", "", errors.New("zero results")
	}
	dest = listprob[0].Destination
	msg = "ini dia list prohibited item dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
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
		return "", err
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

func ExtractKeywords(message string, commonWordsAdd []string) []string {
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

	// Split message into keywords
	keywords := strings.Split(message, " ")

	return keywords
}

func BuildFlexibleRegex(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}

	// Gabungkan kata kunci dengan regex yang memungkinkan urutan apapun
	var regexBuilder strings.Builder
	for _, keyword := range keywords {
		regexBuilder.WriteString("(?=.*\\b" + regexp.QuoteMeta(keyword) + "\\b)")
	}
	regexBuilder.WriteString(".*")
	return regexBuilder.String()
}

func BuildFlexibleRegexWithTypos(keywords []string, db *mongo.Database) string {
	var allKeywords []string
	items, err := atdb.GetAllDoc[Item](db, "prohibited_items_en", bson.M{})
	if err == nil {
		for _, item := range items {
			words := strings.Split(item.ProhibitedItems, " ")
			allKeywords = append(allKeywords, words...)
		}
	}

	var regexBuilder strings.Builder
	for _, keyword := range keywords {
		closestKeyword := findClosestKeyword(keyword, allKeywords)
		regexBuilder.WriteString("(?=.*\\b" + regexp.QuoteMeta(closestKeyword) + "\\b)")
	}
	regexBuilder.WriteString(".*")
	return regexBuilder.String()
}

func findClosestKeyword(keyword string, allKeywords []string) string {
	closestKeyword := keyword
	minDistance := len(keyword) + 1
	for _, kw := range allKeywords {
		distance := smetrics.WagnerFischer(keyword, kw, 1, 1, 2)
		if distance < minDistance {
			minDistance = distance
			closestKeyword = kw
		}
	}
	return closestKeyword
}