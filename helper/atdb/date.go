package atdb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetYesterdayStartEnd() (startDay, endDay primitive.ObjectID) {
	// Hitung tanggal kemarin
	loc, _ := time.LoadLocation("Asia/Jakarta")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999, loc)
	startDay = primitive.NewObjectIDFromTimestamp(startOfDay)
	endDay = primitive.NewObjectIDFromTimestamp(endOfDay)
	return

}
