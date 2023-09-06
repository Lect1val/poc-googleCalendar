package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type App struct {
	Config *oauth2.Config
}

func (app *App) getTokenFromWeb() *oauth2.Token {
	authURL := app.Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := app.Config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

// func (app *App) callbackHandler(c *gin.Context) {
// 	code := c.DefaultQuery("code", "")
// 	if code == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No code provided"})
// 		return
// 	}

// 	token, err := app.Config.Exchange(context.TODO(), code)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
// 		return
// 	}

// 	fmt.Printf("Access token: %s\n", token.AccessToken)
// 	c.JSON(http.StatusOK, gin.H{"message": "Token retrieved successfully"})
// }

func createEvent(srv *calendar.Service) (*calendar.Event, error) {
	event := &calendar.Event{
		Summary:     "test send invite 09-05_1",
		Location:    "The PARQ, 5th Floor",
		Description: "poc test send event",
		Start:       &calendar.EventDateTime{Date: "2023-09-07"},
		End:         &calendar.EventDateTime{Date: "2023-09-08"},
		Attendees: []*calendar.EventAttendee{
			{Email: "example001@gmail.com"},
		},
	}

	newEvent, err := srv.Events.Insert("primary", event).SendUpdates("all").Do()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Event created with ID: %s\n", newEvent.Id)
	return newEvent, nil
}

// func listEvents(srv *calendar.Service, maxResults int64) ([]*calendar.Event, error) {
// 	events, err := srv.Events.List("primary").MaxResults(maxResults).Do()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return events.Items, nil
// }

func listUpcomingEvents(srv *calendar.Service, maxResults int64) ([]*calendar.Event, error) {
	now := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").MaxResults(maxResults).TimeMin(now).Do()
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

// func deleteEvent(srv *calendar.Service, eventId string) error {
// 	return srv.Events.Delete("primary", eventId).Do()
// }

func readEvent(srv *calendar.Service, eventId string) (*calendar.Event, error) {
	return srv.Events.Get("primary", eventId).Do()
}

func updateEvent(srv *calendar.Service, eventId string, event *calendar.Event) (*calendar.Event, error) {
	return srv.Events.Update("primary", eventId, event).SendUpdates("all").Do()
}

func inviteAttendees(srv *calendar.Service, eventId string, newAttendees []*calendar.EventAttendee) (*calendar.Event, error) {
	event, err := readEvent(srv, eventId)
	if err != nil {
		return nil, err
	}

	event.Reminders = &calendar.EventReminders{
		UseDefault: true,
	}

	event.Attendees = append(event.Attendees, newAttendees...)
	return updateEvent(srv, eventId, event)
}

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	app := &App{
		Config: config,
	}

	var token *oauth2.Token
	file, err := os.Open("token.json")
	if err != nil {
		token = app.getTokenFromWeb()
		tokenFile, err := os.OpenFile("token.json", os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Fatalf("Unable to cache oauth token: %v", err)
		}
		defer tokenFile.Close()
		json.NewEncoder(tokenFile).Encode(token)
	} else {
		defer file.Close()
		token = &oauth2.Token{}
		err = json.NewDecoder(file).Decode(token)
		if err != nil {
			token = app.getTokenFromWeb()
			tokenFile, err := os.OpenFile("token.json", os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				log.Fatalf("Unable to cache oauth token: %v", err)
			}
			defer tokenFile.Close()
			json.NewEncoder(tokenFile).Encode(token)
		}
	}

	client := config.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client: %v", err)
	}

	// newEvent, err := createEvent(srv)
	// if err != nil {
	// 	log.Fatalf("Failed to create event: %v", err)
	// }

	// fmt.Printf("Event created: %s\n", newEvent.HtmlLink)

	moreAttendees := []*calendar.EventAttendee{
		{Email: "example001@gmail.com"},
	}

	updatedEvent, err := inviteAttendees(srv, "pqgi1et8sb7pstt6j035vf34gc", moreAttendees)
	if err != nil {
		log.Fatalf("Failed to invite more attendees: %v", err)
	}

	fmt.Printf("Event updated with more attendees: %s\n", updatedEvent.HtmlLink)

	// events, err := listEvents(srv, 10)
	// if err != nil {
	// 	log.Fatalf("Failed to list events: %v", err)
	// }
	// fmt.Println("Upcoming events:")
	// for _, e := range events {
	// 	fmt.Printf("- %s (%s)\n", e.Summary, e.Start.DateTime)
	// }

	events, err := listUpcomingEvents(srv, 20)
	if err != nil {
		log.Fatalf("Failed to list events: %v", err)
	}
	// fmt.Println("Upcoming events:")
	// for _, e := range events {
	// 	summary := "No Title"
	// 	if e.Summary != "" {
	// 		summary = e.Summary
	// 	}

	// 	startTime := "No Start Time"
	// 	if e.Start != nil && e.Start.DateTime != "" {
	// 		startTime = e.Start.DateTime
	// 	}

	// 	fmt.Printf("- %s (%s)\n", summary, startTime)
	// }
	for _, e := range events {
		summary := "No Title"
		if e.Summary != "" {
			summary = e.Summary
		}

		startTime := "No Start Time"
		endTime := "No End Time"
		if e.Start != nil {
			if e.Start.DateTime != "" {
				startTime = e.Start.DateTime
			} else if e.Start.Date != "" {
				startTime = e.Start.Date + " (All day)"
			}
		}
		if e.End != nil {
			if e.End.DateTime != "" {
				endTime = e.End.DateTime
			} else if e.End.Date != "" {
				endTime = e.End.Date + " (All day)"
			}
		}
		eventId := e.Id
		location := e.Location
		description := e.Description
		creator := ""
		if e.Creator != nil {
			creator = e.Creator.Email
		}

		attendees := []string{}
		for _, a := range e.Attendees {
			attendees = append(attendees, a.Email)
		}
		attendeeList := strings.Join(attendees, ", ")

		fmt.Printf("EventID:%s\n Event: %s\nStart Time: %s\nEnd Time: %s\nLocation: %s\nDescription: %s\nCreator: %s\nAttendees: %s\n\n",
			eventId, summary, startTime, endTime, location, description, creator, attendeeList)
	}

	// r := gin.Default()
	// r.GET("/callback", app.callbackHandler)
	// r.Run(":8080")
}
