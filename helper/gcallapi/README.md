# gcallapi

Untuk mendapatkan token pertama kali,juga jika token habis buka ini dan jalankan di lokal:

```go
func main() {
 // Connect to MongoDB
 client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://a98jiu"))
 if err != nil {
  log.Fatal(err)
 }
 defer client.Disconnect(context.TODO())

 db := client.Database("dbname")
 conf, _ := credentialsFromDB(db)
 tok := GetTokenFromWeb(conf)
 saveToken(db, tok)
}

// Request a token from the web, then returns the retrieved token
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
 authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
 fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

 var authCode string
 if _, err := fmt.Scan(&authCode); err != nil {
  panic(fmt.Sprintf("Unable to read authorization code: %v", err))
 }

 tok, err := config.Exchange(context.TODO(), authCode)
 if err != nil {
  panic(fmt.Sprintf("Unable to retrieve token from web: %v", err))
 }
 return tok
}
```

## Cara pakai

Isi model saja

```go
simpleEvent := SimpleEvent{
        Summary:     "Google I/O 2024",
        Location:    "800 Howard St., San Francisco, CA 94103",
        Description: "A chance to hear more about Google's developer products.",
        Date:        "2024-06-14",
        TimeStart:   "09:00:00",
        TimeEnd:     "17:00:00",
        Attendees:   []string{"awangga@gmail.com", "awangga@ulbi.ac.id"},
    }
```

gmail

```go
// Define the email details
    to := "recipient@example.com"
    subject := "Test Email with Attachment"
    body := "This is a test email with attachment."
    attachmentPaths := []string{"path/to/attachment1.pdf", "path/to/attachment2.jpg"}

    // Send the email
    err = gcallapi.SendEmailWithAttachment(db, to, subject, body, attachmentPaths)
    if err != nil {
        log.Fatalf("Error sending email: %v", err)
    }
```

```go
// Define the email details
 to := "recipient@example.com"
 subject := "Test Email"
 body := "This is a test email."

 // Send the email
 err = gcallapi.SendEmail(db, to, subject, body)
 if err != nil {
  log.Fatalf("Error sending email: %v", err)
 }

```

blogger

```go
 // Define the blog details
 blogID := "your-blog-id" // Ganti dengan ID blog Anda
 title := "Test Post"
 content := "This is a test post."
 content := `
  <h1>This is a test post</h1>
  <p>This is a paragraph with <strong>bold</strong> text and <em>italic</em> text.</p>
  <ul>
   <li>Item 1</li>
   <li>Item 2</li>
   <li>Item 3</li>
  </ul>
 `

 // Post to Blogger
 post, err := gcallapi.PostToBlogger(db, blogID, title, content)
 if err != nil {
  log.Fatalf("Error posting to Blogger: %v", err)
 }

 blogID := "your-blog-id"
 postID := "your-post-id"

 err = gcallapi.DeletePostFromBlogger(db, blogID, postID)
 if err != nil {
  log.Fatalf("Failed to delete post: %v", err)
 }

```

drive

```go
fileID := "your-file-id"
 newTitle := "Duplicated File"

 duplicatedFile, err := gcallapi.DuplicateFileInDrive(db, fileID, newTitle)
 if err != nil {
  log.Fatalf("Failed to duplicate file: %v", err)
 }
```

docs

```go
docID := "your-doc-id"
 replacements := map[string]string{
  "oldText1": "newText1",
  "oldText2": "newText2",
 }

 err = gcallapi.ReplaceStringsInDoc(db, docID, replacements)
 if err != nil {
  log.Fatalf("Failed to replace strings in document: %v", err)
 }
```

pdf

```go
docID := "your-doc-id"
 outputFileName := "output.pdf"

 fileID, err := gcallapi.GeneratePDF(db, docID, outputFileName)
 if err != nil {
  log.Fatalf("Failed to generate PDF: %v", err)
 }

```
