package kimseok

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	levenshtein "github.com/texttheater/golang-levenshtein/levenshtein"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Fungsi utama untuk mendapatkan nama negara dari pesan
func GetCountryFromMessage(message string, db *mongo.Database) (negara, msg, collection string, err error) {
    lowerMessage := strings.ToLower(message)
    collection = "prohibited_items_id"
    
    // Mengambil daftar nama negara dalam bahasa Indonesia
    listnegara, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destinasi", collection)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }
    
    // Mencari kecocokan terdekat dalam daftar negara
    negara, distance := GetClosestMatch(lowerMessage, listnegara)
    if distance <= 3 { // Mengizinkan kesalahan ketik kecil (misalnya, hingga 3 karakter)
        msg = strings.ReplaceAll(lowerMessage, strings.ToLower(negara), "")
        msg = strings.TrimSpace(msg)
        return
    }

    // Jika tidak ada kecocokan, coba dengan daftar nama negara dalam bahasa Inggris
    collection = "prohibited_items_en"
    countrylist, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destination", collection)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }

    negara, distance = GetClosestMatch(lowerMessage, countrylist)
    if distance <= 3 { // Mengizinkan kesalahan ketik kecil dalam bahasa Inggris juga
        msg = strings.ReplaceAll(lowerMessage, strings.ToLower(negara), "")
        msg = strings.TrimSpace(msg)
        return
    }

    return
}

// Fungsi untuk mendapatkan barang terlarang dari pesan
func GetProhibitedItemsFromMessage(negara, message string, db *mongo.Database, collectionName string) (bool, string, string, error) {
    var fieldTujuan, fieldBarang string

    switch collectionName {
    case "prohibited_items_id":
        fieldTujuan = "Destinasi"
        fieldBarang = "Barang Terlarang"
    default:
        fieldTujuan = "Destination"
        fieldBarang = "Prohibited Items"
    }

    var msg string
    var additionalMsg string = "☎ Ini dia nih Call Centre Hallo Pos 📞1500161, bukan tempat buat curhat ya Kak! Atau kakak bisa mengirimkan keluh kesalnya ke email kami di\n✉ halopos@posindonesia.co.id"
    var additionalMsgHelp string = "Apa ada lagi yang bisa aku bantu kak? (づ ◕‿◕ )づ"

    if negara != "" {
        msg = "💡 Berikut ini adalah daftar barang yang dilarang dari negara *" + negara + "* Kak:\n"

        filter := bson.M{fieldTujuan: bson.M{"$regex": negara, "$options": "i"}}
        if message != "" {
            msg += "Dengan kata kunci _*" + message + "*_:\n"
            filter[fieldBarang] = bson.M{"$regex": message, "$options": "i"}
        }

        if collectionName == "prohibited_items_id" {
            return processProhibitedItems(db, collectionName, filter, negara, message, msg, additionalMsg, additionalMsgHelp, true)
        }
        return processProhibitedItems(db, collectionName, filter, negara, message, msg, additionalMsg, additionalMsgHelp, false)
    }

    return false, "", "", nil
}

// Fungsi untuk memproses barang terlarang
func processProhibitedItems(db *mongo.Database, collectionName string, filter bson.M, negara, message, msg, additionalMsg, additionalMsgHelp string, isIndonesian bool) (bool, string, string, error) {
    if isIndonesian {
        prohitems, err := atdb.GetAllDoc[[]DestinasiTerlarang](db, collectionName, filter)
        if err != nil {
            return false, "", "", fmt.Errorf("error fetching countries from DB IND: %v", err)
        }

        if len(prohitems) != 0 {
            for i, item := range prohitems {
                msg += strconv.Itoa(i+1) + ". " + item.BarangTerlarang + "\n"
            }
            additionalMsg += "\n" + additionalMsgHelp
            return true, msg, additionalMsg, nil
        }

        return true, "📚 *" + message + "* diperbolehkan untuk dikirim ke negara *" + negara + "* Kak!\n", additionalMsg, nil
    } else {
        prohitems, err := atdb.GetAllDoc[[]DestinationProhibit](db, collectionName, filter)
        if err != nil {
            return false, "", "", fmt.Errorf("error fetching countries from DB ENG: %v", err)
        }

        if len(prohitems) != 0 {
            for i, item := range prohitems {
                msg += strconv.Itoa(i+1) + ". " + item.ProhibitedItems + "\n"
            }
            additionalMsg += "\n" + additionalMsgHelp
            return true, msg, additionalMsg, nil
        }

        return true, "📚 *" + message + "* is allowed to be sent to *" + negara + "* Mastah!\n", additionalMsg, nil
    }
}

// Fungsi untuk mencari kecocokan terdekat dengan pencocokan kabur
func GetClosestMatch(input string, candidates []interface{}) (string, int) {
    input = strings.ToLower(input)
    minDistance := -1
    closestMatch := ""
    for _, candidate := range candidates {
        country := strings.ToLower(candidate.(string))
        distance := levenshtein.DistanceForStrings([]rune(input), []rune(country), levenshtein.DefaultOptions)
        if minDistance == -1 || distance < minDistance {
            minDistance = distance
            closestMatch = candidate.(string)
        }
    }

    return closestMatch, minDistance
}
