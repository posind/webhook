package kimseok

import (
    "fmt"
    "log"
    "strconv"
    "strings"

    "github.com/gocroot/helper/atdb"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

func GetCountryFromMessage(message string, db *mongo.Database) (negara, msg, collection string, err error) {
    lowerMessage := strings.ToLower(message)
    collection = "prohibited_items_id"
    listnegara, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destinasi", collection)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }
    for _, country := range listnegara {
        if strings.Contains(lowerMessage, strings.ToLower(country.(string))) {
            msg = strings.ReplaceAll(lowerMessage, strings.ToLower(country.(string)), "")
            msg = strings.TrimSpace(msg)
            negara = country.(string)
            return
        }
    }

    collection = "prohibited_items_en"
    countrylist, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destination", collection)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }
    for _, country := range countrylist {
        if strings.Contains(lowerMessage, strings.ToLower(country.(string))) {
            msg = strings.ReplaceAll(lowerMessage, strings.ToLower(country.(string)), "")
            msg = strings.TrimSpace(msg)
            negara = country.(string)
            return
        }
	}

	// collection = "max_weight"
    // maxweight, err := atdb.GetAllDistinctDoc(db, bson.M{}, "Destinasi Negara", collection)
    // if err != nil {
    //     log.Printf("Error fetching countries from DB: %v", err)
    //     return
    // }
    // for _, country := range maxweight {
    //     if strings.Contains(lowerMessage, strings.ToLower(country.(string))) {
    //         msg = strings.ReplaceAll(lowerMessage, strings.ToLower(country.(string)), "")
    //         msg = strings.TrimSpace(msg)
    //         negara = country.(string)
    //         return
    //     }
	// }

    return
}

// func GetMaxWeight(negara, message string, db *mongo.Database, collectionName string) (found bool, msg string, err error) {
//     var fieldNegara, fieldKode, fieldWeight string
//     if collectionName == "max_weight" {
//         fieldNegara = "Destinasi Negara"
//         fieldKode = "Kode Negara"
//         fieldWeight = "Berat Per Koli"
//     }
//     if negara != "" {
//         msg = "Daftar barang terlarang dari negara *" + negara + "*:\n"
//         var filter bson.M
//         if message == "" {
//             filter = bson.M{
//                 fieldNegara: bson.M{
//                     "$regex":   negara,
//                     "$options": "i",
//                 },
//             }
//         } else {
//             msg += "Dengan kata kunci _" + message + "_:\n"
//             filter = bson.M{
//                 fieldNegara: bson.M{
//                     "$regex":   negara,
//                     "$options": "i",
//                 },
//                 fieldKode: bson.M{
//                     "$regex":   message,
//                     "$options": "i",
//                 },
//                 fieldWeight: bson.M{
//                     "$regex":   message,
//                     "$options": "i",
//                 },
//             }
//         }
//         if collectionName == "max_weight" {
//             maxitems, errr := atdb.GetAllDoc[[]MaxWeight](db, collectionName, filter)
//             if errr != nil {
//                 log.Printf("Error fetching max weight from DB: %v", errr)
//                 err = fmt.Errorf("error fetching max weight from DB: %v", errr)
//                 return
//             }
//             if len(maxitems) != 0 {
//                 for i, item := range maxitems {
//                     msg += strconv.Itoa(i+1) + ". " + item.BeratPerKoli + "\n"
//                 }
//                 found = true
//             } else if message != "" {
//                 msg += "_tidak ditemukan_\nBerikut berat maksimal per koli untuk negara " + negara + ":\n"
//                 filter = bson.M{
//                     fieldNegara: bson.M{
//                         "$regex":   negara,
//                         "$options": "i",
//                     },
//                 }
//                 maxitems, err = atdb.GetAllDoc[[]MaxWeight](db, collectionName, filter)
//                 if err != nil {
//                     log.Printf("Error fetching max weight from DB: %v", err)
//                     err = fmt.Errorf("error fetching max weight from DB: %v", err)
//                     return
//                 }
//                 if len(maxitems) != 0 {
//                     for i, item := range maxitems {
//                         msg += strconv.Itoa(i+1) + ". " + item.BeratPerKoli + "\n"
//                     }
//                     found = true
//                 }
//             }
//         }
//     }
//     return
// }

//Untuk Func Get Massage
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
	var additionalMsg string = "☎ Ini Nih Call Centre Hallo Pos  📞1500161, bukan tempat buat curhat ya Kak! Atau kakak bisa mengirimkan keluh kesalnya ke email kami di\n✉ halopos@posindonesia.co.id"

    if negara != "" {
        msg = "💡 Berikut ini adalah daftar barang yang dilarang dari negara *" + negara + "* Kak:\n"

        filter := bson.M{fieldTujuan: bson.M{"$regex": negara, "$options": "i"}}
        if message != "" {
            msg += "Dengan kata kunci _*" + message + "*_:\n"
            filter[fieldBarang] = bson.M{"$regex": message, "$options": "i"}
        }

        if collectionName == "prohibited_items_id" {
            return processProhibitedItems(db, collectionName, filter, negara, message, msg, additionalMsg, true)
        }
        return processProhibitedItems(db, collectionName, filter, negara, message, msg, additionalMsg, false)
    }

    return false, "", "", nil
}

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
