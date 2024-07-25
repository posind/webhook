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

// GetProhibitedItems retrieves prohibited items based on the user's message.
func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) string {
	keywords := extractKeywords(Pesan.Message, []string{})
	country, item, err := getCountryAndItemFromKeywords(keywords, db)
	if err != nil {
		return "Error: " + err.Error()
	}

	if country == "" {
		return "Nama negara tidak ada kak di database kita"
	}

	filter := createFilter(country, item, db)
	reply, err := populateList(db, filter, item)
	if err != nil {
		return "📚" + strings.Join(keywords, " ") + "|" + country + " : " + err.Error() + "\n" + filterToString(filter)
	}

	return "📚" + reply
}

// getCountryAndItemFromKeywords identifies the country and item from the provided keywords.
func getCountryAndItemFromKeywords(keywords []string, db *mongo.Database) (string, string, error) {
	var (
		country string
		err     error
	)
	remainingKeywords := make([]string, len(keywords))
	copy(remainingKeywords, keywords)

	for len(remainingKeywords) > 0 {
		remainingMessage := strings.Join(remainingKeywords, " ")
		country, err = getCountryNameLike(db, remainingMessage)
		if err == nil {
			break
		}
		remainingKeywords = remainingKeywords[:len(remainingKeywords)-1]
	}

	if country != "" {
		item := strings.Join(keywords[len(remainingKeywords):], " ")
		return country, item, nil
	}

	return "", "", errors.New("negara tidak ditemukan di pesan")
}

// populateList builds a response message listing prohibited items for the specified country and item.
func populateList(db *mongo.Database, filter bson.M, keyword string) (string, error) {
	listprob, err := atdb.GetAllDoc[Item](db, "prohibited_items_id", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", err
	}
	if len(listprob) == 0 {
		return "Tidak ada prohibited items yang ditemukan", errors.New("zero results")
	}

	dest := listprob[0].Destinasi
	msg := "Ini dia list prohibited item dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.BrangTerlarang + "\n"
	}
	return msg, nil
}

// getCountryNameLike finds a country in the database that matches the provided name (case-insensitive).
func getCountryNameLike(db *mongo.Database, country string) (string, error) {
	filter := bson.M{"Destinasi": bson.M{"$regex": country, "$options": "i"}}
	itemprohb, err := atdb.GetOneDoc[Item](db, "prohibited_items_id", filter)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(itemprohb.Destinasi, "\u00A0", " "), nil
}

// extractKeywords cleans and extracts keywords from a message.
func extractKeywords(message string, commonWordsAdd []string) []string {
	commonWords := []string{"list", "barang", "barang barang", "barang", "mymy"}
	commonWords = append(commonWords, commonWordsAdd...)
	message = strings.ToLower(strings.ReplaceAll(message, "\u00A0", " "))

	for _, word := range commonWords {
		message = strings.ReplaceAll(message, word, "")
	}

	message = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(message, " "))
	return strings.Split(message, " ")
}

// buildFlexibleRegexWithTypos creates a regex pattern to match items allowing for minor typos.
func buildFlexibleRegexWithTypos(keywords []string, db *mongo.Database) string {
	allKeywords := getAllKeywords(db)
	var regexBuilder strings.Builder
	for _, keyword := range keywords {
		closestKeyword := findClosestKeyword(keyword, allKeywords)
		regexBuilder.WriteString("(?=.*\\b" + regexp.QuoteMeta(closestKeyword) + "\\b)")
	}
	regexBuilder.WriteString(".*")
	return regexBuilder.String()
}

// getAllKeywords retrieves all keywords from the database for typo matching.
func getAllKeywords(db *mongo.Database) []string {
	var allKeywords []string
	items, err := atdb.GetAllDoc[Item](db, "prohibited_items_id", bson.M{})
	if err == nil {
		for _, item := range items {
			allKeywords = append(allKeywords, strings.Split(item.BrangTerlarang, " ")...)
		}
	}
	return allKeywords
}

// findClosestKeyword finds the closest keyword from the database to the provided keyword.
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

// createFilter builds a filter for the MongoDB query based on the country and item.
func createFilter(country, item string, db *mongo.Database) bson.M {
	filter := bson.M{"Destinasi": country}
	if item != "" {
		regexPattern := buildFlexibleRegexWithTypos([]string{item}, db)
		filter["Barang Terlarang"] = bson.M{"$regex": regexPattern, "$options": "i"}
	}
	return filter
}

// filterToString converts a BSON filter to a string representation.
func filterToString(filter bson.M) string {
	jsonData, _ := bson.Marshal(filter)
	return string(jsonData)
}