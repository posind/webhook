// blogger.go
package gcallapi

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/blogger/v3"
	"google.golang.org/api/option"
)

// Helper function to create a Blogger service
func createBloggerService(ctx context.Context, db *mongo.Database) (*blogger.Service, error) {
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
	srv, err := blogger.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// Function to post to Blogger
func PostToBlogger(db *mongo.Database, blogID, title, content string) (*blogger.Post, error) {
	ctx := context.Background()

	srv, err := createBloggerService(ctx, db)
	if err != nil {
		return nil, err
	}

	post := &blogger.Post{
		Title:   title,
		Content: content,
	}

	createdPost, err := srv.Posts.Insert(blogID, post).Do()
	if err != nil {
		return nil, err
	}

	return createdPost, nil
}

// Function to check if a post with the same title already exists in Blogger
func PostExistsInBlogger(db *mongo.Database, blogID, title string) (bool, error) {
	ctx := context.Background()

	srv, err := createBloggerService(ctx, db)
	if err != nil {
		return false, err
	}

	call := srv.Posts.List(blogID).Fields("items(title)")
	posts, err := call.Do()
	if err != nil {
		return false, err
	}

	for _, post := range posts.Items {
		if post.Title == title {
			return true, nil
		}
	}

	return false, nil
}

// Function to create a post in Blogger
func CreatePostInBlogger(db *mongo.Database, blogID, title, content string) (*blogger.Post, error) {
	ctx := context.Background()

	exists, err := PostExistsInBlogger(db, blogID, title)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("post with the same title already exists")
	}

	srv, err := createBloggerService(ctx, db)
	if err != nil {
		return nil, err
	}

	post := &blogger.Post{
		Title:   title,
		Content: content,
	}

	newPost, err := srv.Posts.Insert(blogID, post).Do()
	if err != nil {
		return nil, err
	}

	return newPost, nil
}

// Function to delete a post from Blogger
func DeletePostFromBlogger(db *mongo.Database, blogID, postID string) error {
	ctx := context.Background()

	srv, err := createBloggerService(ctx, db)
	if err != nil {
		return err
	}

	err = srv.Posts.Delete(blogID, postID).Do()
	if err != nil {
		return err
	}

	return nil
}
