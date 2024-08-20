package posint

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/kimseok"
	"github.com/whatsauth/itmodel"
	"github.com/xrash/smetrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetProhibitedItems fetches prohibited items based on the message and database
func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	country, _, _, err := kimseok.GetCountryFromMessage(Pesan.Message, db)
	var filter bson.M
	var keyword string
	if err != nil {
		countryandkeyword := ExtractKeywords(Pesan.Message, []string{})
		words := strings.Split(strings.Join(countryandkeyword, " "), " ")
		var key []string
		if country == "" {
			return "Nama negara nya tidak ada di database kita kak!"
		}
		for len(words) > 0 {
			remainingMessage := strings.Join(words, " ")
			country, err = GetCountryNameLike(db, remainingMessage)
			if err == nil {
				break
			}
			lastWord := words[len(words)-1]
			key = append(key, lastWord)
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
		reply, _, err = populateListProhibited(db, filter, keyword)
		reply = "💡" + reply
		if err != nil {
			if err.Error() == "zero results" {
				return " is allowed to send to " + country
			}
			jsonData, _ := bson.Marshal(filter)
			return "💡" + strings.Join(countryandkeyword, " ") + "|" + country + " : " + err.Error() + "\n" + string(jsonData)
		}
		return
	}

	// Integrasi GetMaxWeight
	reply += "\n\n" + GetMaxWeight(Pesan, db)
	return
}

// GetMaxWeight fetches max weight based on the message and database
func GetMaxWeight(Pesan itmodel.IteungMessage, db *mongo.Database) string {
	keywords := ExtractKeywords(Pesan.Message, nil)
	country, item, err := GetCountryAndItemFromKeywords(keywords, db)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	if country == "" {
		return "Nama negaranya tidak ada di database kita kakak :("
	}
	filter := bson.M{"Destinasi Negara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"}}
	if item != "" {
		regexPattern := BuildFlexibleRegexWithTypos([]string{item}, db)
		filter["Kode Negara"] = bson.M{"$regex": regexPattern, "$options": "i"}
	}
	reply, _, err := populateListMaxWeight(db, filter, item)
	if err != nil {
		jsonData, _ := bson.Marshal(filter)
		return fmt.Sprintf("📚%s|%s : %v\n%s", strings.Join(keywords, " "), country, err, string(jsonData))
	}
	return "📚" + reply
}

// GetCountryAndItemFromKeywords determines the country and item from the given keywords
func GetCountryAndItemFromKeywords(keywords []string, db *mongo.Database) (string, string, error) {
	for i := 0; i < len(keywords); i++ {
		stemmedKeyword := kimseok.Stemmer(keywords[i])
		country, err := GetCountryNameLike(db, stemmedKeyword)
		if err == nil {
			item := strings.Join(append(keywords[:i], keywords[i+1:]...), " ")
			return country, item, nil
		}
	}
	return "", "", errors.New("nama negaranya mana kak?")
}

// GetCountryNameLike searches for a country name in the database
func GetCountryNameLike(db *mongo.Database, country string) (string, error) {
	filter := bson.M{
		"Destination": bson.M{"$regex": country, "$options": "i"},
	}
	itemprohb, err := atdb.GetOneDoc[ItemProhibited](db, "prohibited_items_en", filter)
	if err != nil {
		return "", err
	}
	dest := strings.ReplaceAll(itemprohb.Destination, "\u00A0", " ")
	return dest, nil
}

// populateListProhibited creates a list of prohibited items based on the filter
func populateListProhibited(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listprob, err := atdb.GetAllDoc[[]ItemProhibited](db, "prohibited_items_en", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listprob) == 0 {
		return " Tidak ada prohibited items yang ditemukan", "", errors.New("zero results")
	}
	dest = listprob[0].Destination
	msg = " Ini dia list prohibited item dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci: _" + keyword + "_\n"
	}
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.ProhibitedItems + "\n"
	}
	return msg, dest, nil
}

// populateListMaxWeight creates a list of max weight items based on the filter
func populateListMaxWeight(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listmax, err := atdb.GetAllDoc[[]ItemWeight](db, "max_weight", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listmax) == 0 {
		return "Tidak ada berat maksimal per koli yang ditemukan", "", errors.New("zero results")
	}
	dest = listmax[0].DestinasiNegara
	var msgBuilder strings.Builder
	msgBuilder.WriteString(" Ini dia berat maksimal per koli dari negara *" + dest + "*:\n")
	if keyword != "" {
		msgBuilder.WriteString("kata-kunci:_" + keyword + "_\n")
	}
	for i, item := range listmax {
		msgBuilder.WriteString(strconv.Itoa(i+1) + ". Kode Negara: " + item.KodeNegara + ", Berat per Koli: " + item.BeratPerKoli + "\n")
	}
	return msgBuilder.String(), dest, nil
}

// ExtractKeywords extracts meaningful keywords from a message
func ExtractKeywords(message string, commonWordsAdd []string) []string {
	commonWords := []string{"list", "en", "id", "mymy", "berat", "max", "maks"}
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

// BuildFlexibleRegexWithTypos creates a flexible regex that accounts for typos
func BuildFlexibleRegexWithTypos(keywords []string, db *mongo.Database) string {
	var allKeywords []string
	items, err := atdb.GetAllDoc[[]ItemWeight](db, "max_weight", bson.M{})
	if err != nil {
		return ""
	}
	for _, item := range items {
		words := strings.Split(item.KodeNegara, " ")
		allKeywords = append(allKeywords, words...)
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
	const insertionCost, deletionCost, substitutionCost = 1, 1, 2
	closestKeyword := keyword
	minDistance := len(keyword) + 1
	for _, kw := range allKeywords {
		distance := smetrics.WagnerFischer(keyword, kw, insertionCost, deletionCost, substitutionCost)
		if distance < minDistance {
			minDistance = distance
			closestKeyword = kw
		}
	}
	return closestKeyword
}
