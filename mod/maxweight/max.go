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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetMaxWeight fetches max weight based on the message and database
func GetMaxWeight(Pesan itmodel.IteungMessage, db *mongo.Database) string {
    // Extract keywords from the message
    keywords := ExtractKeywords(Pesan.Message, nil)
    
    // Get country and item from the extracted keywords
    country, item, err := GetCountryAndItemFromKeywords(keywords, db)
    if err != nil {
        return fmt.Sprintf("Error: %v", err)
    }

    // Handle case where country is not found
    if country == "" {
        return "Nama negaranya tidak ada di database kita kakak :("
    }

    // Define filter for searching in collections
    filter := bson.M{"Destinasi Negara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"}}
    if item != "" {
        filter["Kode Negara"] = bson.M{"$regex": item, "$options": "i"}
    }

    // Search in "max_weight" collection
    reply, _, err := populateList(db, "max_weight", filter, item)
    if err != nil {
        // If no data found in "max_weight", search in "max_weight_id"
        reply, _, err = populateList(db, "max_weight_id", filter, item)
        if err != nil {
            jsonData, _ := bson.Marshal(filter)
            return fmt.Sprintf("📚%s|%s : %v\n%s", strings.Join(keywords, " "), country, err, string(jsonData))
        }
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
// GetCountryNameLike searches for a country name in the database, first in "max_weight", then in "max_weight_id" if not found
func GetCountryNameLike(db *mongo.Database, country string) (string, error) {
    // Define the filter
    filter := bson.M{
        "Destinasi Negara": bson.M{"$regex": kimseok.Stemmer(country), "$options": "i"},
    }

    // Try to find in "max_weight" collection
    maxw, err := atdb.GetOneDoc[Item](db, "max_weight", filter)
    if err == nil {
        // Data found in "max_weight"
        dest := strings.ReplaceAll(maxw.DestinasiNegara, "\u00A0", " ")
        return dest, nil
    }

    // If not found in "max_weight", try to find in "max_weight_id"
    maxwID, err := atdb.GetOneDoc[Item](db, "max_weight_id", filter)
    if err != nil {
        return "", err
    }
    
    dest := strings.ReplaceAll(maxwID.DestinasiNegara, "\u00A0", " ")
    return dest, nil
}


// populateList creates a list of items based on the filter and collection
func populateList(db *mongo.Database, collectionName string, filter bson.M, keyword string) (string, string, error) {
    listmax, err := atdb.GetAllDoc[[]Item](db, collectionName, filter)
    if err != nil {
        return "Terdapat kesalahan pada GetAllDoc", "", err
    }
    if len(listmax) == 0 {
        return "Tidak ada berat maksimal per koli yang ditemukan", "", errors.New("zero results")
    }
    dest := listmax[0].DestinasiNegara
    var msg strings.Builder
    msg.WriteString(" Ini dia berat maksimal per koli dari negara *" + dest + "*:\n")
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
    commonWords := []string{"berat", "max", "maks", "mymy"}
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
