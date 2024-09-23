package lms

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetNewCookie mengirim request dan mengembalikan XSRF-TOKEN dan laravel_session yang baru
func GetNewCookie(xsrfToken string, laravelSession string, db *mongo.Database) (string, string, string, error) {
	// Membuat cookie jar untuk menangkap cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", "", "", fmt.Errorf("error creating cookie jar: %w", err)
	}

	client := &http.Client{
		Jar: jar,
	}
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return "", "", "", fmt.Errorf("error get profile db: %w", err)
	}

	// Membuat request
	req, err := http.NewRequest("GET", profile.URLCookie, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("error creating request: %w", err)
	}

	// Menambahkan header ke request
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-US,en;q=0.7")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", fmt.Sprintf("XSRF-TOKEN=%s; laravel_session=%s", xsrfToken, laravelSession))
	req.Header.Set("Host", "pamongdesa.id")
	req.Header.Set("Referer", "https://pamongdesa.id/admin/user")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", "\"Not/A)Brand\";v=\"8\", \"Chromium\";v=\"126\", \"Brave\";v=\"126\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Windows\"")

	// Mengirim request
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Membaca response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("error reading response body: %w", err)
	}
	content := string(respBody)
	// Regex untuk mencari token
	re := regexp.MustCompile(`eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`)
	bearer := re.FindString(content)

	// Menangkap set cookies dari response header
	var newXSRFToken, newLaravelSession string
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case "XSRF-TOKEN":
			newXSRFToken = cookie.Value
		case "laravel_session":
			newLaravelSession = cookie.Value
		}
	}

	if newXSRFToken == "" || newLaravelSession == "" {
		return "", "", bearer, fmt.Errorf("could not find new XSRF-TOKEN or laravel_session in response cookies")
	}

	return newXSRFToken, newLaravelSession, bearer, nil
}
