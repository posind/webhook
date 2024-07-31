package alloweditems

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

// GetAllowedItems fetches and returns the list of allowed items for a specified country
func GetAllowedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
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
			regexPattern := BuildFlexibleRegexWithTypos(ExtractKeywords(keyword, []string{}), db)
			filter = bson.M{
				"Destination":      country,
				"Allowed Items": bson.M{"$regex": regexPattern, "$options": "i"},
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
		regexPattern := BuildFlexibleRegexWithTypos(keywords, db)
		filter = bson.M{
			"Destination":      country,
			"Allowed Items": bson.M{"$regex": regexPattern, "$options": "i"},
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

// populateList fetches and constructs the allowed items list based on the given filter
func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listallowed, err := atdb.GetAllDoc[Allowed_Items](db, "allowed_items", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listallowed) == 0 {
		return "Tidak ada allowed items yang ditemukan", "", errors.New("zero results")
	}
	dest = listallowed[0].Destination
	msg = "Ini dia list allowed item dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
	for i, alloweditem := range listallowed {
		msg += strconv.Itoa(i+1) + ". " + alloweditem.AllowedItems + "\n"
	}
	return
}

func GetCountryNameLike(db *mongo.Database, country string) (dest string, err error) {
	filter := bson.M{
		"Destination": bson.M{"$regex": country, "$options": "i"},
	}
	itemallowed, err := atdb.GetOneDoc[Allowed_Items](db, "allowed_items", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(itemallowed.Destination, "\u00A0", " ")
	return
}

// GetCountryFromMessage extracts the country name from a message string
func GetCountryFromMessage(message string, db *mongo.Database) (country string, err error) {
	// Ubah pesan menjadi huruf kecil
	lowerMessage := strings.ToLower(message)
	// Mengganti non-breaking space dengan spasi biasa
	lowerMessage = strings.ReplaceAll(lowerMessage, "\u00A0", " ")
	// Hapus spasi berlebih
	lowerMessage = strings.TrimSpace(lowerMessage)
	lowerMessage = regexp.MustCompile(`\s+`).ReplaceAllString(lowerMessage, " ")
	// Mendapatkan nama negara
	countries, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destination", "allowed_items")
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

// ExtractKeywords removes common words from a message and returns the remaining keywords
func ExtractKeywords(message string, commonWordsAdd []string) []string {
	// Daftar kata umum yang mungkin ingin dihilangkan
	commonWords := []string{"list", "allowed", "items", "item", "mymy"}
	
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
	keywords := strings.Split(message, " ")
	return keywords
}

// BuildFlexibleRegex creates a regex pattern for flexible keyword matching
func BuildFlexibleRegex(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}
	var regexBuilder strings.Builder
	for _, keyword := range keywords {
		regexBuilder.WriteString("(?=.*\\b" + regexp.QuoteMeta(keyword) + "\\b)")
	}
	regexBuilder.WriteString(".*")
	return regexBuilder.String()
}

// BuildFlexibleRegexWithTypos creates a regex pattern that accounts for typos
func BuildFlexibleRegexWithTypos(keywords []string, db *mongo.Database) string {
	var allKeywords []string
	items, err := atdb.GetAllDoc[Allowed_Items](db, "allowed_items", bson.M{})
	if err == nil {
		for _, item := range items {
			words := strings.Split(item.AllowedItems, " ")
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

// findClosestKeyword finds the closest matching keyword from a list based on edit distance
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
