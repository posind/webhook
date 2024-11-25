package report

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/raykov/gofpdf"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPDFandMDMeeting(db *mongo.Database, projectName string) (base64Str, joinMD string, err error) {
	filter := CreateFilterMeetingYesterday(projectName, true)
	laporanDocs, err := atdb.GetAllDoc[[]model.Laporan](db, "uxlaporan", filter) //CreateFilterMeetingYesterday(projectName)
	if err != nil {
		return
	}
	if len(laporanDocs) == 0 {
		return
	}
	// Buat PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)

	// Menambahkan fungsi footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15) // Posisi footer dari bawah halaman
		pdf.SetFont("Arial", "I", 8)
		pageNo := pdf.PageNo() // Mendapatkan nomor halaman
		pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d", pageNo), "", 0, "C", false, 0, "")
	})

	for i, laporan := range laporanDocs {
		// Tambahkan halaman baru
		pdf.AddPage()
		pdf.SetFont("Arial", "UB", 16)
		judul := "Risalah Pertemuan-" + strconv.Itoa(i+1)
		pdf.MultiCell(
			0,     // Lebar: 0 berarti lebar otomatis
			10,    // Tinggi baris
			judul, // Teks
			"",    // Batas kiri
			"",    // Batas kanan
			false, // Aligment horizontal
		)
		joinMD = joinMD + "# " + judul + "\n"
		// Tambahkan teks ke PDF
		pdf.SetFont("Arial", "B", 12)
		pdf.MultiCell(
			0,                         // Lebar: 0 berarti lebar otomatis
			5,                         // Tinggi baris
			laporan.MeetEvent.Summary, // Teks
			"",                        // Batas kiri
			"",                        // Batas kanan
			false,                     // Aligment horizontal
		)
		joinMD = joinMD + "## " + laporan.MeetEvent.Summary + "\n"
		pdf.SetFont("Arial", "I", 12)
		pdf.MultiCell(
			0,                          // Lebar: 0 berarti lebar otomatis
			5,                          // Tinggi baris
			"Notula: "+laporan.Petugas, // Teks
			"",                         // Batas kiri
			"",                         // Batas kanan
			false,                      // Aligment horizontal
		)
		joinMD = joinMD + "Notula: " + laporan.Petugas + "\n"
		pdf.MultiCell(
			0, // Lebar: 0 berarti lebar otomatis
			5, // Tinggi baris
			"Waktu: "+laporan.MeetEvent.Date+" ("+laporan.MeetEvent.TimeStart+" - "+laporan.MeetEvent.TimeEnd+")", // Teks
			"",    // Batas kiri
			"",    // Batas kanan
			false, // Aligment horizontal
		)
		joinMD = joinMD + "Waktu: " + laporan.MeetEvent.Date + " (" + laporan.MeetEvent.TimeStart + " - " + laporan.MeetEvent.TimeEnd + ")" + "\n"
		pdf.MultiCell(
			0,                                     // Lebar: 0 berarti lebar otomatis
			5,                                     // Tinggi baris
			"Lokasi: "+laporan.MeetEvent.Location, // Teks
			"",                                    // Batas kiri
			"",                                    // Batas kanan
			false,                                 // Aligment horizontal
		)
		joinMD = joinMD + "Lokasi: " + laporan.MeetEvent.Location + "\n"
		pdf.MultiCell(
			0,         // Lebar: 0 berarti lebar otomatis
			5,         // Tinggi baris
			"Agenda:", // Teks
			"",        // Batas kiri
			"",        // Batas kanan
			false,     // Aligment horizontal
		)
		joinMD = joinMD + "### Agenda" + "\n"
		pdf.MultiCell(
			0,              // Lebar: 0 berarti lebar otomatis
			5,              // Tinggi baris
			laporan.Solusi, // Teks
			"",             // Batas kiri
			"",             // Batas kanan
			false,          // Aligment horizontal
		)
		joinMD = joinMD + laporan.Solusi + "\n"
		pdf.SetFont("Arial", "UB", 12)
		pdf.MultiCell(
			0,         // Lebar: 0 berarti lebar otomatis
			5,         // Tinggi baris
			"Risalah", // Teks
			"",        // Batas kiri
			"",        // Batas kanan
			false,     // Aligment horizontal
		)
		joinMD = joinMD + "### Risalah" + "\n"
		pdf.SetFont("Arial", "", 12)
		pdf.MultiCell(
			0,                // Lebar: 0 berarti lebar otomatis
			5,                // Tinggi baris
			laporan.Komentar, // Teks
			"",               // Batas kiri
			"",               // Batas kanan
			false,            // Aligment horizontal
		)
		joinMD = joinMD + laporan.Komentar
	}

	// Simpan PDF ke file sementara
	tempFile := projectName
	err = pdf.OutputFileAndClose(tempFile)
	if err != nil {
		return
	}

	// Baca file PDF dan konversi ke base64
	fileData, err := ioutil.ReadFile(tempFile)
	if err != nil {
		return
	}

	base64Str = base64.StdEncoding.EncodeToString(fileData)

	// Hapus file sementara
	err = os.Remove(tempFile)
	if err != nil {
		return
	}

	return

}
