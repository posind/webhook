package kimseok

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/gocroot/helper/atdb"
	// "github.com/gocroot/mod/helpdesk"
	// "github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetCursorFromRegex(db *mongo.Database, patttern string) (cursor *mongo.Cursor, err error) {
	filter := bson.M{"question": primitive.Regex{Pattern: patttern, Options: "i"}}
	cursor, err = db.Collection("qna").Find(context.TODO(), filter)
	return
}

func GetCursorFromString(db *mongo.Database, question string) (cursor *mongo.Cursor, err error) {
	filter := bson.M{"question": question}
	cursor, err = db.Collection("qna").Find(context.TODO(), filter)
	return
}

func GetRandomFromQnASlice(qnas []Datasets) Datasets {
	// Inisialisasi sumber acak dengan waktu saat ini
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Pilih elemen acak dari slice
	randomIndex := rng.Intn(len(qnas))
	return qnas[randomIndex]
}

func GetQnAfromSliceWithJaro(q string, qnas []Datasets) (dt Datasets) {
	var score float64
	for _, qna := range qnas {
		//fmt.Println(data)
		str2 := qna.Question
		scorex := jaroWinkler(q, str2)
		if score < scorex {
			dt = qna
			score = scorex
		}

	}
	return

}

// // balasan jika tidak ditemukan key word
// func GetMessage(Profile itmodel.Profile, msg itmodel.IteungMessage, botname string, db *mongo.Database) string {
// 	//check apakah ada permintaan operator masuk
// 	reply, err := helpdesk.PenugasanOperator(Profile, msg, db)
// 	if err != nil {
// 		return err.Error()
// 	}
// 	//jika tidak ada di db komplain lanjut ke selanjutnya
// 	if reply == "" {
// 		dt, err := QueriesDataRegexpALL(db, msg.Message)
// 		if err != nil {
// 			return err.Error()
// 		}
// 		reply = strings.TrimSpace(dt.Answer)

// 	}
// 	return reply
// }

func QueriesDataRegexpALL(db *mongo.Database, queries string) (dest Datasets, err error) {
	// Kata akhiran imbuhan
	// queries = SeparateSuffixMu(queries)

	// Ubah ke kata dasar
	queries = Stemmer(queries)
	filter := bson.M{"question": queries}
	qnas, err := atdb.GetAllDoc[[]Datasets](db, "qna", filter)
	if err != nil {
		return
	}
	if len(qnas) > 0 {
		// Flatten qnas if it's [][]Datasets
		flatQnas := flattenDatasets(qnas)
		dest = GetRandomFromQnASlice(flatQnas)
		return
	}

	wordsdepan := strings.Fields(queries) // Tokenization
	wordsbelakang := wordsdepan
	// Pencarian dengan pengurangan kata dari belakang
	for len(wordsdepan) > 0 || len(wordsbelakang) > 0 {
		// Join remaining elements back into a string for wordsdepan
		filter := bson.M{"question": primitive.Regex{Pattern: strings.Join(wordsdepan, " "), Options: "i"}}
		qnas, err = atdb.GetAllDoc[[]Datasets](db, "qna", filter)
		if err != nil {
			return
		}
		if len(qnas) > 0 {
			flatQnas := flattenDatasets(qnas)
			dest = GetQnAfromSliceWithJaro(queries, flatQnas)
			return
		}

		// Join remaining elements back into a string for wordsbelakang
		filter = bson.M{"question": primitive.Regex{Pattern: strings.Join(wordsbelakang, " "), Options: "i"}}
		qnas, err = atdb.GetAllDoc[[]Datasets](db, "qna", filter)
		if err != nil {
			return
		}
		if len(qnas) > 0 {
			flatQnas := flattenDatasets(qnas)
			dest = GetQnAfromSliceWithJaro(queries, flatQnas)
			return
		}

		// Remove the last element
		wordsdepan = wordsdepan[:len(wordsdepan)-1]
		// Remove element pertama
		wordsbelakang = wordsbelakang[1:]
	}

	return
}

// Helper function to flatten a [][]Datasets into []Datasets
func flattenDatasets(qnas [][]Datasets) []Datasets {
	var result []Datasets
	for _, innerSlice := range qnas {
		result = append(result, innerSlice...)
	}
	return result
}
