package alloweditemsen

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"github.com/xrash/smetrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAllowedItems fetches and returns the list of allowed items for a specified country
func GetAllowedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	keywords := ExtractKeywords(Pesan.Message, nil)
	country, item, err := GetCountryAndItemFromKeywords(keywords, db)
	if err != nil {
		return "Error: " + err.Error()
	}

	if country == "" {
		return "Nama negaranya tidak ada di database kita kakak"
	}

	filter := bson.M{"Destination": country}
	if item != "" {
		regexPattern := BuildFlexibleRegexWithTypos([]string{item}, db)
		filter["Allowed Items"] = bson.M{"$regex": regexPattern, "$options": "i"}
	}

	reply, _, err = populateList(db, filter, item)
	reply = "📚" + reply
	if err != nil {
		jsonData, _ := bson.Marshal(filter)
		return "📚" + strings.Join(keywords, " ") + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
	}
	return
}

// GetCountryAndItemFromKeywords determines the country and item from the given keywords
func GetCountryAndItemFromKeywords(keywords []string, db *mongo.Database) (country, item string, err error) {
	for i := 0; i < len(keywords); i++ {
		country, err = GetCountryNameLike(db, keywords[i])
		if err == nil {
			item = strings.Join(append(keywords[:i], keywords[i+1:]...), " ")
			return
		}
	}

	err = errors.New("nama negaranya mana kak?")
	return
}

// GetCountryNameLike searches for a country name in the database
func GetCountryNameLike(db *mongo.Database, country string) (dest string, err error) {
	filter := bson.M{
		"Destination": primitive.Regex{Pattern: country, Options: "i"},
	}
	itemallow, err := atdb.GetOneDoc[Item](db, "allowed_items", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(itemallow.Destination, "\u00A0", " ")
	return
}

// GetItemNameLike searches for an item name in the database
func GetItemNameLike(db *mongo.Database, item string) (dest string, err error) {
	filter := bson.M{
		"Barang yang Dibolehkan": bson.M{"$regex": item, "$options": "i"},
	}
	itemallow, err := atdb.GetOneDoc[Item](db, "allowed_items", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(itemallow.AllowedItems, "\u00A0", " ")
	return
}

// populateList creates a list of prohibited items based on the filter
func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listallow, err := atdb.GetAllDoc[Item](db, "allowed_items", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listallow) == 0 {
		return "Tidak ada barang yang Dibolehkan yang ditemukan", "", errors.New("zero results")
	}
	dest = listallow[0].Destination
	msg = "Ini dia list barang yang Dibolehkan dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
	for i, allowitem := range listallow {
		msg += strconv.Itoa(i+1) + ". " + allowitem.AllowedItems + "\n"
	}
	return
}

// ExtractKeywords extracts meaningful keywords from a message
func ExtractKeywords(message string, commonWordsAdd []string) []string {
	commonWords := []string{"list", "allowed", "allow", "items", "item", "mymy"}
	commonWords = append(commonWords, commonWordsAdd...)
	message = strings.ToLower(message)
	message = strings.ReplaceAll(message, "\u00A0", " ")

	for _, word := range commonWords {
		word = strings.ToLower(strings.ReplaceAll(word, "\u00A0", " "))
		message = strings.ReplaceAll(message, word, "")
	}

	message = strings.TrimSpace(message)
	message = regexp.MustCompile(`\s+`).ReplaceAllString(message, " ")
	keywords := strings.Split(message, " ")

	if len(keywords) > 2 {
		keywords = keywords[:2]
	}

	return keywords
}

// BuildFlexibleRegex constructs a regex pattern that matches all given keywords
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

// BuildFlexibleRegexWithTypos creates a flexible regex that accounts for typos
func BuildFlexibleRegexWithTypos(keywords []string, db *mongo.Database) string {
	var allKeywords []string
	items, err := atdb.GetAllDoc[Item](db, "allowed_items", bson.M{})
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

// findClosestKeyword finds the closest match for a keyword from a list of known words
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
