package gcallapi

import (
	"context"
	"io"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Helper function to create a Drive service
func createDriveService(ctx context.Context, db *mongo.Database) (*drive.Service, error) {
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
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// Function to duplicate a file in Google Drive
func DuplicateFileInDrive(db *mongo.Database, fileID, newTitle string) (*drive.File, error) {
	ctx := context.Background()

	srv, err := createDriveService(ctx, db)
	if err != nil {
		return nil, err
	}

	// Retrieve the file metadata
	originalFile, err := srv.Files.Get(fileID).Fields("id", "name", "mimeType", "parents").Do()
	if err != nil {
		return nil, err
	}

	// Create a copy of the file
	copy := &drive.File{
		Name:    newTitle,
		Parents: originalFile.Parents,
	}

	duplicatedFile, err := srv.Files.Copy(fileID, copy).Do()
	if err != nil {
		return nil, err
	}

	return duplicatedFile, nil
}

// Function to generate a PDF from a Google Doc and save it to Google Drive
func GeneratePDF(db *mongo.Database, docID, outputFileName string) (string, error) {
	ctx := context.Background()

	srv, err := createDriveService(ctx, db)
	if err != nil {
		return "", err
	}

	// Export the Google Doc to PDF
	res, err := srv.Files.Export(docID, "application/pdf").Download()
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Create a temporary file to store the PDF
	outFile, err := os.Create(outputFileName)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		return "", err
	}

	// Upload the PDF file to Google Drive
	f := &drive.File{Name: outputFileName}
	file, err := srv.Files.Create(f).Media(outFile).Do()
	if err != nil {
		return "", err
	}

	return file.Id, nil
}
