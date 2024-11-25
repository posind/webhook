// email.go
package gcallapi

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Function to create a Gmail service
func createGmailService(ctx context.Context, db *mongo.Database) (*gmail.Service, error) {
	// Retrieve OAuth2 config from DB
	config, err := credentialsFromDB(db)
	if err != nil {
		return nil, err
	}

	// Retrieve token from DB
	token, err := tokenFromDB(db)
	if err != nil {
		return nil, err
	}

	// Refresh the token if it has expired
	if token.Expiry.Before(time.Now()) {
		token, err = refreshToken(config, token)
		if err != nil {
			return nil, err
		}
		err = saveToken(db, token)
		if err != nil {
			return nil, err
		}
	}

	client := config.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// Helper function to create email body
func CreateEmail(to string, subject string, body string) ([]byte, error) {
	var msg strings.Builder
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n" + body)

	return []byte(msg.String()), nil
}

// Function to send an email with attachment
func SendEmailWithAttachment(db *mongo.Database, to string, subject string, body string, attachmentPaths []string) error {
	ctx := context.Background()

	srv, err := createGmailService(ctx, db)
	if err != nil {
		return err
	}

	// Create email
	var msg strings.Builder
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/mixed; boundary=\"boundary\"\r\n")
	msg.WriteString("\r\n--boundary\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n" + body + "\r\n")

	for _, path := range attachmentPaths {
		content, err := readFile(path)
		if err != nil {
			return err
		}

		encoded := base64.StdEncoding.EncodeToString(content)
		msg.WriteString("\r\n--boundary\r\n")
		msg.WriteString("Content-Type: application/octet-stream; name=\"" + path + "\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString("Content-Disposition: attachment; filename=\"" + path + "\"\r\n")
		msg.WriteString("\r\n" + encoded + "\r\n")
	}

	msg.WriteString("--boundary--")

	var message gmail.Message
	message.Raw = base64.URLEncoding.EncodeToString([]byte(msg.String()))

	_, err = srv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return err
	}

	return nil
}

// Function to send an email without attachment
func SendEmail(db *mongo.Database, to string, subject string, body string) error {
	ctx := context.Background()

	srv, err := createGmailService(ctx, db)
	if err != nil {
		return err
	}

	// Create email
	var msg strings.Builder
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n" + body + "\r\n")

	var message gmail.Message
	message.Raw = base64.URLEncoding.EncodeToString([]byte(msg.String()))

	_, err = srv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return err
	}

	return nil
}

// Helper function to read a file and return its content
func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)

	bytesRead, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buffer[:bytesRead], nil
}
