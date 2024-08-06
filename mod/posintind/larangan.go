package posintid

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

// GetProhibitedItems fetches prohibited items based on the message and database
func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	keywords := ExtractKeywords(Pesan.Message, nil)
	country, item, err := GetCountryAndItemFromKeywords(keywords, db)
	if err != nil {
		return "Error: " + err.Error()
	}

	if country == "" {
		return "Nama negaranya tidak ada di database kita kakak"
	}

	filter := bson.M{"Destinasi": country}
	if item != "" {
		regexPattern := BuildFlexibleRegexWithTypos([]string{item}, db)
		filter["Barang Terlarang"] = bson.M{"$regex": regexPattern, "$options": "i"}
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

	err = errors.New("nama negaranya mana kak")
	return
}

// GetCountryNameLike searches for a country name in the database
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

// GetItemNameLike searches for an item name in the database
func GetItemNameLike(db *mongo.Database, item string) (dest string, err error) {
	filter := bson.M{
		"Barang Terlarang": bson.M{"$regex": item, "$options": "i"},
	}
	itemprohb, err := atdb.GetOneDoc[Item](db, "prohibited_items_id", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(itemprohb.BarangTerlarang, "\u00A0", " ")
	return
}

// populateList creates a list of prohibited items based on the filter
func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listprob, err := atdb.GetAllDoc[Item](db, "prohibited_items_id", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listprob) == 0 {
		return "Tidak ada barang terlarang yang ditemukan", "", errors.New("zero results")
	}
	dest = listprob[0].Destinasi
	msg = "Ini dia list barang terlarang dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.BarangTerlarang + "\n"
	}
	return
}

// ExtractKeywords extracts meaningful keywords from a message
func ExtractKeywords(message string, commonWordsAdd []string) []string {
	commonWords := []string{"list", "id", "indo", "mymy"}
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
	items, err := atdb.GetAllDoc[Item](db, "prohibited_items_id", bson.M{})
	if err == nil {
		for _, item := range items {
			words := strings.Split(item.BarangTerlarang, " ")
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
