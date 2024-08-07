package maxweight

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

// GetMaxWeight fetches max weight based on the message and database
func GetMaxWeight(Pesan itmodel.IteungMessage, db *mongo.Database) string {
    keywords := ExtractKeywords(Pesan.Message, nil)
    country, item, err := GetCountryAndItemFromKeywords(keywords, db)
    if err != nil {
        return fmt.Sprintf("Error: %v", err)
    }

    if country == "" {
        return "Nama negaranya tidak ada di database kita kakak"
    }

    filter := bson.M{"DestinasiNegara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"}}
    if item != "" {
        regexPattern := BuildFlexibleRegexWithTypos([]string{item}, db)
        filter["KodeNegara"] = bson.M{"$regex": regexPattern, "$options": "i"}
    }

    reply, _, err := populateList(db, filter, item)
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
    return "", "", errors.New("nama negaranya mana kak")
}

// GetCountryNameLike searches for a country name in the database
func GetCountryNameLike(db *mongo.Database, country string) (string, error) {
    filter := bson.M{
        "DestinasiNegara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"},
    }
    maxw, err := atdb.GetOneDoc[Item](db, "max_weight", filter)
    if err != nil {
        return "", err
    }
    dest := strings.ReplaceAll(maxw.DestinasiNegara, "\u00A0", " ")
    return dest, nil
}

// GetItemNameLike searches for an item name in the database
func GetItemNameLike(db *mongo.Database, item string) (string, string, error) {
    filter := bson.M{
        "BarangTerlarang": bson.M{"$regex": item, "$options": "i"},
    }
    maxwei, err := atdb.GetOneDoc[Item](db, "max_weight", filter)
    if err != nil {
        return "", "", err
    }
    kodeNegara := strings.ReplaceAll(maxwei.KodeNegara, "\u00A0", " ")
    beratPerKoli := strings.ReplaceAll(maxwei.BeratPerKoli, "\u00A0", " ")
    return kodeNegara, beratPerKoli, nil
}

// populateList creates a list of items based on the filter
func populateList(db *mongo.Database, filter bson.M, keyword string) (string, string, error) {
    listmax, err := atdb.GetAllDoc[Item](db, "max_weight", filter)
    if err != nil {
        return "Terdapat kesalahan pada GetAllDoc", "", err
    }
    if len(listmax) == 0 {
        return "Tidak ada barang terlarang yang ditemukan", "", errors.New("zero results")
    }
    dest := listmax[0].DestinasiNegara
    var msg strings.Builder
    msg.WriteString("Ini dia daftar dari negara *" + dest + "*:\n")
    if keyword != "" {
        msg.WriteString("kata-kunci:_" + keyword + "_\n")
    }
    for i, item := range listmax {
        msg.WriteString(strconv.Itoa(i+1) + ". Kode Negara: " + item.KodeNegara + ", Berat per Koli: " + item.BeratPerKoli + "\n")
    }
    return msg.String(), dest, nil
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
