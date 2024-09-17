package kimseok

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetCountryFromMessageTele(message string, db *mongo.Database) (negara, msg, collection string, err error) {
    // Daftar kata umum yang bisa diabaikan
    commonWords := []string{"bisa", "berikan", "informasi", "terkait", "pengiriman", "ke", "negara", "barang", "apakah"}

    // Ubah pesan menjadi huruf kecil dan bersihkan dari kata-kata umum
    lowerMessage := strings.ToLower(message)
    for _, commonWord := range commonWords {
        lowerMessage = strings.ReplaceAll(lowerMessage, commonWord, "")
    }
    lowerMessage = strings.TrimSpace(lowerMessage)

    // Coba cari negara dengan menggunakan regex
    cursor, err := GetCursorFromRegex(db, lowerMessage)
    if err != nil {
        log.Printf("Error fetching countries from DB using regex: %v", err)
        return
    }
    defer cursor.Close(context.TODO())

    // Proses hasil pencarian negara
    for cursor.Next(context.TODO()) {
        var country bson.M
        if err = cursor.Decode(&country); err != nil {
            log.Printf("Error decoding country: %v", err)
            continue
        }

        // Ambil nama negara
        negara = country["Destinasi"].(string) // Sesuaikan field ini dengan yang ada di database
        msg = strings.ReplaceAll(lowerMessage, strings.ToLower(negara), "")
        msg = strings.TrimSpace(msg)
        collection = "prohibited_items_id"
        return
    }

    // Jika tidak ditemukan di koleksi pertama, coba koleksi lainnya (bahasa Inggris)
    collection = "prohibited_items_en"
    cursor, err = GetCursorFromRegex(db, lowerMessage)
    if err != nil {
        log.Printf("Error fetching countries from DB: %v", err)
        return
    }
    defer cursor.Close(context.TODO())

    for cursor.Next(context.TODO()) {
        var country bson.M
        if err = cursor.Decode(&country); err != nil {
            log.Printf("Error decoding country: %v", err)
            continue
        }

        // Ambil nama negara
        negara = country["Destination"].(string) // Sesuaikan dengan field di database
        msg = strings.ReplaceAll(lowerMessage, strings.ToLower(negara), "")
        msg = strings.TrimSpace(msg)
        return
    }

    return
}

// Fungsi untuk memproses pesan dari Telegram
func GetProhibitedItemsFromMessageTele(negara, message string, db *mongo.Database, collectionName string) (bool, string, error) {
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
    // var additionalMsg string = "Ada yang bisa aku bantu lagi ga kak? \n (づ ◕‿◕ )づ"
    // var additionalMsg string = "☎ Ini dia nih Call Centre Hallo Pos  📞1500161, bukan tempat buat curhat ya Kak! Atau kakak bisa mengirimkan keluh kesalnya ke email kami di\n✉ halopos@posindonesia.co.id"

    if negara != "" {
        msg = "💡Ini dia nih kak, barang yang dilarang dari negara *" + negara + "*:\n"

        filter := bson.M{fieldTujuan: bson.M{"$regex": negara, "$options": "i"}}
        if message != "" {
            msg += "dengan kategori *" + message + "*:\n"
            filter[fieldBarang] = bson.M{"$regex": message, "$options": "i"}
        }

        if collectionName == "prohibited_items_id" {
            return processProhibitedItemsTele(db, collectionName, filter, negara, message, msg, true)
        }
        return processProhibitedItemsTele(db, collectionName, filter, negara, message, msg, false)
    }

    return false, "", nil
}

// Fungsi untuk memproses data dari MongoDB (untuk Telegram)
func processProhibitedItemsTele(db *mongo.Database, collectionName string, filter bson.M, negara, message, msg string, isIndonesian bool) (bool, string, error) {
    if isIndonesian {
        prohitems, err := atdb.GetAllDoc[[]DestinasiTerlarang](db, collectionName, filter)
        if err != nil {
            return false, "", fmt.Errorf("error fetching countries from DB IND: %v", err)
        }

        if len(prohitems) != 0 {
            for i, item := range prohitems {
                // Penambahan ke string 'msg' tanpa masalah tipe
                msg += strconv.Itoa(i+1) + ". " + item.BarangTerlarang + "\n"
            }
            return true, msg, nil
        }

        return true, "📚 *" + message + "* diperbolehkan untuk dikirim ke negara *" + negara + "* Kak!\n", nil
    } else {
        prohitems, err := atdb.GetAllDoc[[]DestinationProhibit](db, collectionName, filter)
        if err != nil {
            return false, "", fmt.Errorf("error fetching countries from DB ENG: %v", err)
        }

        if len(prohitems) != 0 {
            for i, item := range prohitems {
                // Penambahan ke string 'msg' tanpa masalah tipe
                msg += strconv.Itoa(i+1) + ". " + item.ProhibitedItems + "\n"
            }
            return true, msg, nil
        }

        return true, "📚 *" + message + "* is allowed to be sent to *" + negara + "* Mastah!\n", nil
    }
}
