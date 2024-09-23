package gcallapi

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
)

// Retrieve credentials.json from MongoDB
func credentialsFromDB(db *mongo.Database) (*oauth2.Config, error) {
	collection := db.Collection("credentials")
	var credentialRecord CredentialRecord
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&credentialRecord)
	if err != nil {
		return nil, err
	}

	if len(credentialRecord.RedirectURIs) == 0 {
		return nil, errors.New("no redirect URIs found in credentials")
	}

	config := &oauth2.Config{
		ClientID:     credentialRecord.ClientID,
		ClientSecret: credentialRecord.ClientSecret,
		Scopes:       credentialRecord.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  credentialRecord.AuthURI,
			TokenURL: credentialRecord.TokenURI,
		},
		RedirectURL: credentialRecord.RedirectURIs[0], // Using the first redirect URI
	}

	return config, nil
}

// Retrieve a token from MongoDB
func tokenFromDB(db *mongo.Database) (*oauth2.Token, error) {
	collection := db.Collection("tokens")
	var tokenRecord CredentialRecord
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&tokenRecord)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenRecord.Token,
		RefreshToken: tokenRecord.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       tokenRecord.Expiry,
	}
	if tokenRecord.Token == "" {
		return nil, errors.New("token tidak ada")
	}

	return token, nil
}

// Saves a token to MongoDB
func saveToken(db *mongo.Database, token *oauth2.Token) error {
	collection := db.Collection("tokens")
	tokenRecord := bson.M{
		"token":         token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	}

	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{},
		bson.M{"$set": tokenRecord},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return nil
}

// Refresh the token using the refresh token
func refreshToken(config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	ts := config.TokenSource(context.Background(), token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}
