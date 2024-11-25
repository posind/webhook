package report

import (
	"strconv"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// menambah poin untuk presensi
func TambahPoinTasklistbyPhoneNumber(db *mongo.Database, phonenumber string, project model.Project, poin float64, activity string) (res *mongo.UpdateResult, err error) {
	usr, err := atdb.GetOneDoc[model.Userdomyikado](db, "user", bson.M{"phonenumber": phonenumber})
	if err != nil {
		return
	}
	usr.Poin = usr.Poin + poin
	res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"phonenumber": phonenumber}, usr)
	if err != nil {
		return
	}

	logpoin := LogPoin{
		UserID:           usr.ID,
		Name:             usr.Name,
		PhoneNumber:      usr.PhoneNumber,
		Email:            usr.Email,
		Poin:             poin,
		ProjectID:        project.ID,
		ProjectName:      project.Name,
		ProjectWAGroupID: project.WAGroupID,
		Activity:         activity,
	}
	//memasukkan detil task ke dalam log
	taskdoing, err := atdb.GetOneLatestDoc[TaskList](db, "taskdoing", bson.M{"phonenumber": usr.PhoneNumber})
	if err == nil {
		taskdoing.Poin = taskdoing.Poin + poin
		res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"_id": taskdoing.ID}, usr)
		if err == nil {
			logpoin.TaskID = taskdoing.ID
			logpoin.Task = taskdoing.Task
			logpoin.LaporanID = taskdoing.LaporanID
		}
	}
	_, err = atdb.InsertOneDoc(db, "logpoin", logpoin)
	if err != nil {
		return
	}

	return

}

// menambah poin untuk presensi
func TambahPoinPresensibyPhoneNumber(db *mongo.Database, phonenumber string, lokasi string, poin float64, token, api, activity string) (res *mongo.UpdateResult, err error) {
	usr, err := atdb.GetOneDoc[model.Userdomyikado](db, "user", bson.M{"phonenumber": phonenumber})
	if err != nil {
		return
	}
	usr.Poin = usr.Poin + poin
	res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"phonenumber": phonenumber}, usr)
	if err != nil {
		return
	}
	logpoin := LogPoin{
		UserID:      usr.ID,
		Name:        usr.Name,
		PhoneNumber: usr.PhoneNumber,
		Email:       usr.Email,
		Poin:        poin,
		Lokasi:      lokasi,
		Activity:    activity,
		Location:    lokasi,
	}
	//memasukkan detil task ke dalam log
	taskdoing, err := atdb.GetOneLatestDoc[TaskList](db, "taskdoing", bson.M{"phonenumber": usr.PhoneNumber})
	if err == nil {
		taskdoing.Poin = taskdoing.Poin + poin
		res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"_id": taskdoing.ID}, usr)
		if err == nil {
			logpoin.TaskID = taskdoing.ID
			logpoin.Task = taskdoing.Task
			logpoin.LaporanID = taskdoing.LaporanID
			logpoin.ProjectID = taskdoing.ProjectID
			logpoin.ProjectName = taskdoing.ProjectName
			logpoin.ProjectWAGroupID = taskdoing.ProjectWAGroupID
		}
		if taskdoing.ProjectWAGroupID != "" {
			msg := "*Presensi*\n" + usr.Name + "(" + strconv.Itoa(int(usr.Poin)) + ") - " + usr.PhoneNumber + "\nLokasi: " + lokasi + "\nPoin: " + strconv.Itoa(int(poin))
			dt := &whatsauth.TextMessage{
				To:       taskdoing.ProjectWAGroupID,
				IsGroup:  true,
				Messages: msg,
			}
			_, _, err = atapi.PostStructWithToken[model.Response]("Token", token, dt, api)
			if err != nil {
				return
			}
		}
	}
	_, err = atdb.InsertOneDoc(db, "logpoin", logpoin)
	if err != nil {
		return
	}

	return

}

// menambah poin untuk laporan
func TambahPoinLaporanbyPhoneNumber(db *mongo.Database, prj model.Project, phonenumber string, poin float64, activity string) (res *mongo.UpdateResult, err error) {
	usr, err := atdb.GetOneDoc[model.Userdomyikado](db, "user", bson.M{"phonenumber": phonenumber})
	if err != nil {
		return
	}
	usr.Poin = usr.Poin + poin
	res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"phonenumber": phonenumber}, usr)
	if err != nil {
		return
	}
	logpoin := LogPoin{
		UserID:      usr.ID,
		Name:        usr.Name,
		PhoneNumber: usr.PhoneNumber,
		Email:       usr.Email,
		ProjectID:   prj.ID,
		ProjectName: prj.Name,
		Poin:        poin,
		Activity:    activity,
	}
	//memasukkan detil task ke dalam log
	taskdoing, err := atdb.GetOneLatestDoc[TaskList](db, "taskdoing", bson.M{"phonenumber": usr.PhoneNumber})
	if err == nil {
		taskdoing.Poin = taskdoing.Poin + poin
		res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"_id": taskdoing.ID}, usr)
		if err == nil {
			logpoin.TaskID = taskdoing.ID
			logpoin.Task = taskdoing.Task
			logpoin.LaporanID = taskdoing.LaporanID
			logpoin.ProjectID = prj.ID
			logpoin.ProjectName = prj.Name
			logpoin.ProjectWAGroupID = prj.WAGroupID
		}
	}
	_, err = atdb.InsertOneDoc(db, "logpoin", logpoin)
	if err != nil {
		return
	}

	return

}

func KurangPoinUserbyPhoneNumber(db *mongo.Database, phonenumber string, poin float64) (res *mongo.UpdateResult, err error) {
	usr, err := atdb.GetOneDoc[model.Userdomyikado](db, "user", bson.M{"phonenumber": phonenumber})
	if err != nil {
		return
	}
	usr.Poin = usr.Poin - poin
	res, err = atdb.ReplaceOneDoc(db, "user", bson.M{"phonenumber": phonenumber}, usr)
	if err != nil {
		return
	}
	return

}

func TambahPoinPushRepobyGithubUsername(db *mongo.Database, prj model.Project, report model.PushReport, poin float64) (usr model.Userdomyikado, err error) {
	usr, err = atdb.GetOneDoc[model.Userdomyikado](db, "user", bson.M{"githubusername": report.Username})
	if err != nil {
		return
	}
	usr.Poin = usr.Poin + poin
	_, err = atdb.ReplaceOneDoc(db, "user", bson.M{"githubusername": report.Username}, usr)
	if err != nil {
		return
	}
	logpoin := LogPoin{
		UserID:      usr.ID,
		Name:        usr.Name,
		PhoneNumber: usr.PhoneNumber,
		Email:       usr.Email,
		ProjectID:   prj.ID,
		ProjectName: prj.Name,
		Poin:        poin,
		Activity:    "Push Repo",
		URL:         report.Repo,
		Info:        report.Ref,
		Detail:      report.Message,
	}
	//memasukkan detil task ke dalam log
	taskdoing, err := atdb.GetOneLatestDoc[TaskList](db, "taskdoing", bson.M{"phonenumber": usr.PhoneNumber})
	if err == nil {
		taskdoing.Poin = taskdoing.Poin + poin
		_, err = atdb.ReplaceOneDoc(db, "user", bson.M{"_id": taskdoing.ID}, usr)
		if err == nil {
			logpoin.TaskID = taskdoing.ID
			logpoin.Task = taskdoing.Task
			logpoin.LaporanID = taskdoing.LaporanID
			logpoin.ProjectID = prj.ID
			logpoin.ProjectName = prj.Name
			logpoin.ProjectWAGroupID = prj.WAGroupID
		}
	}
	_, err = atdb.InsertOneDoc(db, "logpoin", logpoin)
	if err != nil {
		return
	}
	return

}

func TambahPoinPushRepobyGithubEmail(db *mongo.Database, prj model.Project, report model.PushReport, poin float64) (usr model.Userdomyikado, err error) {
	usr, err = atdb.GetOneDoc[model.Userdomyikado](db, "user", bson.M{"email": report.Email})
	if err != nil {
		return
	}
	usr.Poin = usr.Poin + poin
	_, err = atdb.ReplaceOneDoc(db, "user", bson.M{"email": report.Email}, usr)
	if err != nil {
		return
	}
	logpoin := LogPoin{
		UserID:      usr.ID,
		Name:        usr.Name,
		PhoneNumber: usr.PhoneNumber,
		Email:       usr.Email,
		ProjectID:   prj.ID,
		ProjectName: prj.Name,
		Poin:        poin,
		Activity:    "Push Repo",
		URL:         report.Repo,
		Info:        report.Ref,
		Detail:      report.Message,
	}
	//memasukkan detil task ke dalam log
	taskdoing, err := atdb.GetOneLatestDoc[TaskList](db, "taskdoing", bson.M{"phonenumber": usr.PhoneNumber})
	if err == nil {
		taskdoing.Poin = taskdoing.Poin + poin
		_, err = atdb.ReplaceOneDoc(db, "user", bson.M{"_id": taskdoing.ID}, usr)
		if err == nil {
			logpoin.TaskID = taskdoing.ID
			logpoin.Task = taskdoing.Task
			logpoin.LaporanID = taskdoing.LaporanID
			logpoin.ProjectID = prj.ID
			logpoin.ProjectName = prj.Name
			logpoin.ProjectWAGroupID = prj.WAGroupID
		}
	}
	_, err = atdb.InsertOneDoc(db, "logpoin", logpoin)
	if err != nil {
		return
	}
	return
}
