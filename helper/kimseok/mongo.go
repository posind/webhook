package kimseok

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/RadhiFadlillah/go-sastrawi"
	"github.com/kimseokgis/backend-ai/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/aiteung/atdb"

	"go.mongodb.org/mongo-driver/mongo"
)

func SetConnection() *mongo.Database {
	var DBmongoinfo = atdb.DBInfo{
		DBString: os.Getenv("MONGOSTRING"),
		DBName:   "Webhook",
	}
	return atdb.MongoConnect(DBmongoinfo)
}

func InsertUserdata(MongoConn *mongo.Database, username, email, password, passwordhash string) (InsertedID interface{}) {
	req := new(model.User)
	req.Username = username
	req.Email = email
	req.Password = password
	req.PasswordHash = passwordhash
	return InsertOneDoc(MongoConn, "users", req)
}

func InsertOneDoc(db *mongo.Database, collection string, doc interface{}) (insertedID interface{}) {
	insertResult, err := db.Collection(collection).InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Printf("InsertOneDoc: %v\n", err)
	}
	return insertResult.InsertedID
}

// Password
func HashPass(passwordhash string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwordhash), 14)
	return string(bytes), err
}

func CheckPasswordHash(passwordhash, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordhash))
	return err == nil
}

func IsPasswordValid(mongoconn *mongo.Database, userdata model.User) bool {
	filter := bson.M{
		"$or": []bson.M{
			{"username": userdata.Username},
			{"email": userdata.Email},
		},
	}

	var res model.User
	err := mongoconn.Collection("users").FindOne(context.TODO(), filter).Decode(&res)

	if err == nil {
		return CheckPasswordHash(userdata.PasswordHash, res.PasswordHash)
	}
	return false
}

// Get User
func FindUserByUsername(conn *mongo.Database, username string) (*model.User, error) {
	var user model.User
	collection := conn.Collection("users")
	filter := bson.M{"username": username}
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengguna tidak ditemukan")
		}
		return nil, err
	}
	return &user, nil
}

func FindAllUsers(conn *mongo.Database) ([]model.User, error) {
	var users []model.User
	collection := conn.Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var user model.User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func QueriesDataRegexp(db *mongo.Database, ctx context.Context, queries string) (dest model.Datasets, err error) {
	filter := bson.M{"questions": primitive.Regex{Pattern: queries, Options: "i"}}
	err = db.Collection("datasets").FindOne(ctx, filter).Decode(&dest)

	if err != nil && err != mongo.ErrNoDocuments {
		return dest, err
	}

	return dest, err
}

func QueriesSecret(db *mongo.Database, ctx context.Context, secret string) (dest model.Secrets, err error) {
	filter := bson.M{"secret_token": primitive.Regex{Pattern: secret, Options: "i"}}
	err = db.Collection("Secret").FindOne(ctx, filter).Decode(&dest)

	if err != nil && err != mongo.ErrNoDocuments {
		return dest, err
	}

	return dest, err
}

func QueriesALL(db *mongo.Database, ctx context.Context) (dest []model.Datasets, err error) {
	filter := bson.M{}
	cursor, err := db.Collection("datasets").Find(ctx, filter)

	if err != nil && err != mongo.ErrNoDocuments {
		return dest, err
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var data model.Datasets
		err := cursor.Decode(&data)
		if err != nil {
			return nil, err
		}
		dest = append(dest, data)
	}

	return dest, err
}

func Stemmer(Sentences string) (newString string) {
	dictionary := sastrawi.DefaultDictionary()
	stemmer := sastrawi.NewStemmer(dictionary)
	for _, word := range sastrawi.Tokenize(Sentences) {
		//if word != "i" { //menghilangkan i singkatan dari iteung
		newString = newString + " " + stemmer.Stem(word)
		//}
	}
	return strings.TrimSpace(newString)
}

// Fungsi untuk memisahkan kata dengan imbuhan "mu" di akhir
func SeparateSuffixMu(word string) string {
	// Regex untuk mendeteksi kata dengan imbuhan "mu" di akhir
	re := regexp.MustCompile(`(\w+)(mu)$`)

	// Cek apakah kata cocok dengan regex
	if re.MatchString(word) {
		// Ganti "mu" dengan " kamu"
		return re.ReplaceAllString(word, "$1 kamu")
	}

	// Jika tidak ada imbuhan "mu", kembalikan kata asli
	return word
}

// Fungsi untuk memisahkan kata dengan imbuhan "mu" di akhir
func SeparateSuffixNya(word string) string {
	// Regex untuk mendeteksi kata dengan imbuhan "mu" di akhir
	re := regexp.MustCompile(`(\w+)(nya)$`)

	// Cek apakah kata cocok dengan regex
	if re.MatchString(word) {
		// Ganti "mu" dengan " kamu"
		return re.ReplaceAllString(word, "$1 dia")
	}

	// Jika tidak ada imbuhan "mu", kembalikan kata asli
	return word
}
