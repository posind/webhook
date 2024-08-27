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

func GetProhibitedItemsFromMessage(negara, message string, db *mongo.Database, collectionName string) (found bool, msg string, err error) {
    var fieldTujuan, fieldBarang string
	if collectionName == "prohibited_items_id" {
		fieldTujuan = "Destinasi"
		fieldBarang = "Barang Terlarang"
	} else {
		fieldTujuan = "Destination"
		fieldBarang = "Prohibited Items"
	}
	// Mengambil data prohibited items dari database MongoDB
	if negara != "" {
		msg = "💡 Daftar barang terlarang dari negara *" + negara + "*:\n"
		// Membuat filter untuk pencarian nama negara dengan regex yang tidak case-sensitive
		var filter bson.M
		if message == "" { // tidak ada kata kunci hanya nama negara saja di pesan
			filter = bson.M{
				fieldTujuan: bson.M{
					"$regex":   negara,
					"$options": "i",
				},
			}
		} else { // ada nama negara dan kata kunci
			msg += "Dengan kata kunci _*" + message + "*_:\n"
			filter = bson.M{
				fieldTujuan: bson.M{
					"$regex":   negara,
					"$options": "i",
				},
				fieldBarang: bson.M{
					"$regex":   message,
					"$options": "i",
				},
			}
		}
		//dapatkan dan parsing hasil
		if collectionName == "prohibited_items_id" {
			prohitems, errr := atdb.GetAllDoc[[]DestinasiTerlarang](db, collectionName, filter)
			if errr != nil {
				err = fmt.Errorf("error fetching countries from DB IND: %v", errr)
				return
			}
			//check apakah hasilnya kosong
			if len(prohitems) != 0 {
				for i, item := range prohitems {
					msg += strconv.Itoa(i+1) + ". " + item.BarangTerlarang + "\n"
				}
				found = true
			} else {
                filter = bson.M{"Destinasi": negara}
                return true, "📚 " + message + " diperbolehkan untuk dikirim ke negara " + negara, nil
			}
		} else {
			prohitems, errr := atdb.GetAllDoc[[]DestinationProhibit](db, collectionName, filter)
			if errr != nil {
				err = fmt.Errorf("error fetching countries from DB ENG: %v", errr)
				return
			}
			if len(prohitems) != 0 {
				for i, item := range prohitems {
					msg += strconv.Itoa(i+1) + ". " + item.ProhibitedItems + "\n"
				}
				found = true
			} else {
                filter = bson.M{"Destination": negara}
                return true, "📚 " + message + " is allowed to be send to " + negara, nil
			}
		}
	}
	return
}
