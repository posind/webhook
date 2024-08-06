package maxweight

import (
	"errors"
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

// GetMaxWeight fetches max weight based on the message and database
func GetMaxWeight(Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	keywords := ExtractKeywords(Pesan.Message, nil)
	country, item, err := GetCountryAndItemFromKeywords(keywords, db)
	if err != nil {
		return "Error: " + err.Error()
	}

	if country == "" {
		return "Nama negaranya tidak ada di database kita kakak"
	}

	filter := bson.M{"DestinasiNegara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"}}
	if item != "" {
		regexPattern := BuildFlexibleRegexWithTypos([]string{item}, db)
		filter["KodeNegara"] = bson.M{"$regex": regexPattern, "$options": "i"}
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
		stemmedKeyword := kimseok.Stemmer(keywords[i])
		country, err = GetCountryNameLike(db, stemmedKeyword)
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
		"DestinasiNegara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"},
	}
	maxw, err := atdb.GetOneDoc[Item](db, "max_weight", filter)
	if err != nil {
		return
	}
	dest = strings.ReplaceAll(maxw.DestinasiNegara, "\u00A0", " ")
	return
}

// GetItemNameLike searches for an item name in the database
func GetItemNameLike(db *mongo.Database, item string) (kodeNegara, beratPerKoli string, err error) {
	filter := bson.M{
		"BarangTerlarang": bson.M{"$regex": item, "$options": "i"},
	}
	maxwei, err := atdb.GetOneDoc[Item](db, "max_weight", filter)
	if err != nil {
		return
	}
	kodeNegara = strings.ReplaceAll(maxwei.KodeNegara, "\u00A0", " ")
	beratPerKoli = strings.ReplaceAll(maxwei.BeratPerKoli, "\u00A0", " ")
	return
}

// populateList creates a list of items based on the filter
func populateList(db *mongo.Database, filter bson.M, keyword string) (msg, dest string, err error) {
	listmax, err := atdb.GetAllDoc[Item](db, "max_weight", filter)
	if err != nil {
		return "Terdapat kesalahan pada GetAllDoc", "", err
	}
	if len(listmax) == 0 {
		return "Tidak ada barang terlarang yang ditemukan", "", errors.New("zero results")
	}
	dest = listmax[0].DestinasiNegara
	msg = "Ini dia daftar dari negara *" + dest + "*:\n"
	if keyword != "" {
		msg += "kata-kunci:_" + keyword + "_\n"
	}
	for i, item := range listmax {
		msg += strconv.Itoa(i+1) + ". Kode Negara: " + item.KodeNegara + ", Berat per Koli: " + item.BeratPerKoli + "\n"
	}
	return
}

// ExtractKeywords extracts meaningful keywords from a message
func ExtractKeywords(message string, commonWordsAdd []string) []string {
	commonWords := []string{"berat", "max", "maks", "weight", "mymy"}
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
	items, err := atdb.GetAllDoc[Item](db, "max_weight", bson.M{})
	if err != nil {
		// Handle error, possibly return an empty regex or some default
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
