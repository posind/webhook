package report

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Fungsi untuk membuat filter
func CreateFilterMeetingYesterday(projectName string, ismeeting bool) bson.M {
	return bson.M{
		"_id": bson.M{
			"$gte": primitive.NewObjectIDFromTimestamp(GetDateKemarin()),
			"$lt":  primitive.NewObjectIDFromTimestamp(GetDateKemarin().Add(24 * time.Hour)),
		},
		"project.name": projectName,
		"meetid":       bson.M{"$exists": ismeeting}, // Kondisi MeetID tidak kosong
	}
}
