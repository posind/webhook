package atdb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddDocToArray[T any](db *mongo.Database, collection string, ObjectID primitive.ObjectID, arrayname string, newDoc T) (result *mongo.UpdateResult, err error) {
	filter := bson.M{"_id": ObjectID}
	update := bson.M{
		"$push": bson.M{arrayname: newDoc},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err = db.Collection(collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return
	}
	return
}

// memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
func DeleteDocFromArray[T any](db *mongo.Database, collection string, ObjectID primitive.ObjectID, arrayname string, memberToDelete T) (result *mongo.UpdateResult, err error) {
	filter := bson.M{"_id": ObjectID}
	update := bson.M{
		"$pull": bson.M{arrayname: memberToDelete},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err = db.Collection(collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return
	}
	return
}

// Menentukan kondisi filter untuk elemen array yang ingin diupdate
// filterCondition := bson.M{"githubusername": "awangga"}
// Nilai baru yang ingin diupdate
// updatedFields := bson.M{"poin": 50}
func EditDocInArray(db *mongo.Database, collection string, ObjectID primitive.ObjectID, arrayname string, filterCondition bson.M, updatedFields bson.M) (result *mongo.UpdateResult, err error) {
	// Membuat filter untuk menemukan dokumen dengan ID dan elemen array yang sesuai dengan kondisi filter
	filter := bson.M{
		"_id": ObjectID,
	}

	// Menambahkan kondisi filter untuk elemen dalam array
	for key, value := range filterCondition {
		filter[fmt.Sprintf("%s.%s", arrayname, key)] = value
	}

	// Membuat update map untuk mengubah elemen array yang sesuai
	updateMap := bson.M{}
	for key, value := range updatedFields {
		updateMap[fmt.Sprintf("%s.$.%s", arrayname, key)] = value
	}

	// Membuat update menggunakan $set
	update := bson.M{
		"$set": updateMap,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Melakukan update pada dokumen yang sesuai dengan filter
	result, err = db.Collection(collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return
	}
	return
}