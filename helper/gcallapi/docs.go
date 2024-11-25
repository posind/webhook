package gcallapi

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// Helper function to create a Docs service
func createDocsService(ctx context.Context, db *mongo.Database) (*docs.Service, error) {
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
	srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// Function to replace multiple strings in a Google Doc
func ReplaceStringsInDoc(db *mongo.Database, docID string, replacements map[string]string) error {
	ctx := context.Background()

	srv, err := createDocsService(ctx, db)
	if err != nil {
		return err
	}

	var requests []*docs.Request

	// Iterate through replacements map and create replace requests
	for oldText, newText := range replacements {
		requests = append(requests, &docs.Request{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					Text:      oldText,
					MatchCase: true,
				},
				ReplaceText: newText,
			},
		})
	}

	// Batch update the document with replace requests
	_, err = srv.Documents.BatchUpdate(docID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return err
	}

	return nil
}
