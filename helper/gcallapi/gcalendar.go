package gcallapi

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func HandlerCalendar(db *mongo.Database, simpleEvent SimpleEvent) (*calendar.Event, error) {
	ctx := context.Background()

	// Retrieve OAuth2 config from DB
	config, err := credentialsFromDB(db)
	if err != nil {
		return nil, err
	}

	// Retrieve token from DB
	token, err := tokenFromDB(db)
	if err != nil {
		// If token not found or expired, get a new one from web and save it
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
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	startDateTime := fmt.Sprintf("%sT%s+07:00", simpleEvent.Date, simpleEvent.TimeStart)
	endDateTime := fmt.Sprintf("%sT%s+07:00", simpleEvent.Date, simpleEvent.TimeEnd)

	var attendees []*calendar.EventAttendee
	for _, email := range simpleEvent.Attendees {
		attendees = append(attendees, &calendar.EventAttendee{Email: email})
	}

	event := &calendar.Event{
		Summary:     simpleEvent.Summary,
		Location:    simpleEvent.Location,
		Description: simpleEvent.Description,
		Start: &calendar.EventDateTime{
			DateTime: startDateTime,
			TimeZone: "Asia/Jakarta",
		},
		End: &calendar.EventDateTime{
			DateTime: endDateTime,
			TimeZone: "Asia/Jakarta",
		},
		Attendees: attendees,
	}
	if len(simpleEvent.Attachments) > 0 {
		event.Attachments = make([]*calendar.EventAttachment, len(simpleEvent.Attachments))
		for i, attachment := range simpleEvent.Attachments {
			event.Attachments[i] = &calendar.EventAttachment{
				FileUrl:  attachment.FileUrl,
				MimeType: attachment.MimeType,
				Title:    attachment.Title,
			}
		}

	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		return nil, err
	}

	return event, nil
}
