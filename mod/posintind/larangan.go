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

func GetProhibitedItems(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	keywords := ExtractKeywords(Pesan.Message, []string{})
	country, item, err := GetCountryAndItemFromKeywords(keywords, db)
	if err != nil {
		return "Error: " + err.Error()
	}
	
	if country == "" {
		return "Nama negara tidak ada kak di database kita"
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

func GetCountryAndItemFromKeywords(keywords []string, db *mongo.Database) (country, item string, err error) {
	remainingKeywords := make([]string, len(keywords))
	copy(remainingKeywords, keywords)

	for len(remainingKeywords) > 0 {
		remainingMessage := strings.Join(remainingKeywords, " ")
		country, err = GetCountryNameLike(db, remainingMessage)
		if err == nil {
			break
		}
		remainingKeywords = remainingKeywords[:len(remainingKeywords)-1]
	}

	if country != "" {
		item = strings.Join(keywords[len(remainingKeywords):], " ")
	} else {
		err = errors.New("negara tidak ditemukan di pesan")
	}

	return
}

func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listprob, err := atdb.GetAllDoc[Item](db, "prohibited_items_id", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listprob) == 0 {
		return "Tidak ada prohibited items yang ditemukan", "", errors.New("zero results")
	}
	dest = listprob[0].Destinasi
	msg = "Ini dia list prohibited item dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
	for i, probitem := range listprob {
		msg += strconv.Itoa(i+1) + ". " + probitem.BrangTerlarang + "\n"
	}
	return
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

func ExtractKeywords(message string, commonWordsAdd []string) []string {
	commonWords := []string{"list", "prohibited", "items", "item", "mymy"}
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
	return keywords
}

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

func BuildFlexibleRegexWithTypos(keywords []string, db *mongo.Database) string {
	var allKeywords []string
	items, err := atdb.GetAllDoc[Item](db, "prohibited_items_id", bson.M{})
	if err == nil {
		for _, item := range items {
			words := strings.Split(item.BrangTerlarang, " ")
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