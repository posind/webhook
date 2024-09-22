package kimseok

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "sort" // Import the sort package
    "github.com/gocroot/helper/atdb"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

func GetCountryFromMessage(message string, db *mongo.Database) (negara, msg, collection string, err error) {
    lowerMessage := strings.ToLower(message)
    collection = "prohibited_items_id"

    // Fetch list of countries as []string instead of []interface{}
    listnegara, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destinasi", collection)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }

    // Convert to []string for proper processing
    countries := make([]string, len(listnegara))
    for i, country := range listnegara {
        countries[i] = country.(string)
    }

    // Sort countries by length (longest first) to prioritize multi-word countries
    sortedCountries := sortCountriesByLength(countries)

    for _, country := range sortedCountries {
        lowerCountry := strings.ToLower(country)

        // Check if the exact multi-word country is in the message
        if strings.Contains(lowerMessage, lowerCountry) {
            msg = strings.ReplaceAll(lowerMessage, lowerCountry, "")
            msg = strings.TrimSpace(msg)
            negara = country
            return
        }
    }

    // Repeat the process for the English collection
    collection = "prohibited_items_en"
    countrylist, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destination", collection)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }

    // Convert to []string
    countries = make([]string, len(countrylist))
    for i, country := range countrylist {
        countries[i] = country.(string)
    }

    // Sort countries by length again for English names
    sortedCountries = sortCountriesByLength(countries)

    for _, country := range sortedCountries {
        lowerCountry := strings.ToLower(country)

        if strings.Contains(lowerMessage, lowerCountry) {
            msg = strings.ReplaceAll(lowerMessage, lowerCountry, "")
            msg = strings.TrimSpace(msg)
            negara = country
            return
        }
    }

    return
}

// Helper function to sort countries by length (longest first)
func sortCountriesByLength(countries []string) []string {
    // Sort countries by descending length to match multi-word countries first
    sort.Slice(countries, func(i, j int) bool {
        return len(countries[i]) > len(countries[j])
    })
    return countries
}

//Untuk Func Get Massage WA
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
    var additionalMsg string = "Ada yang bisa aku bantu lagi ga kak? \n (づ ◕‿◕ )づ"

    if negara != "" {
        msg = "💡Ini dia nih kak, barang yang *dilarang* dari negara *" + negara + "*:\n"

        // Adjust regex to enforce exact or close matching for the country name
        filter := bson.M{fieldTujuan: bson.M{"$regex": "^" + negara + "$", "$options": "i"}}
        if message != "" {
            msg += "dengan kategori *" + message + "*:\n"
            filter[fieldBarang] = bson.M{"$regex": message, "$options": "i"}
        }

        if collectionName == "prohibited_items_id" {
            return processProhibitedItems(db, collectionName, filter, negara, message, msg, additionalMsg, true)
        }
        return processProhibitedItems(db, collectionName, filter, negara, message, msg, additionalMsg, false)
    }

    return false, "", "", nil
}

//Untuk WA
func processProhibitedItems(db *mongo.Database, collectionName string, filter bson.M, negara, message, msg, additionalMsg string, isIndonesian bool) (bool, string, string, error) {
    if isIndonesian {
        prohitems, err := atdb.GetAllDoc[[]DestinasiTerlarang](db, collectionName, filter)
        if err != nil {
            return false, "", "", fmt.Errorf("error fetching countries from DB IND: %v", err)
        }

        if len(prohitems) != 0 {
            for i, item := range prohitems {
                msg += strconv.Itoa(i+1) + ". " + item.BarangTerlarang + "\n"
            }
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
            return true, msg, additionalMsg, nil
        }

        return true, "📚 *" + message + "* is allowed to be sent to *" + negara + "* Mastah!\n", additionalMsg, nil
    }
}
