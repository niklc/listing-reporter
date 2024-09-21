package reporter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type EmailClient struct {
	gmailSvc *gmail.Service
}

type Email struct {
	To      string
	Listing Listing
}

func NewEmailClient(configFile []byte, tokenFile []byte) (*EmailClient, error) {
	config, err := google.ConfigFromJSON(configFile, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create config from credentials: %w", err)
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(tokenFile, token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	ctx := context.Background()

	client := config.Client(ctx, token)

	svc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return &EmailClient{gmailSvc: svc}, nil
}

func (e *EmailClient) SendListing(email Email) error {
	listing := email.Listing

	rows := []map[string]string{
		{"url": fmt.Sprintf("<a href=\"%s\">%s</a>", listing.Url, listing.Url)},
		{"image": fmt.Sprintf("<img src=\"%s\">", listing.Img)},
		{"price": fmt.Sprintf("%.2f", listing.Price)},
		{"price/m2": fmt.Sprintf("%.2f", listing.PricePerM2)},
		{"title": listing.Title},
		{"street": listing.Street},
		{"rooms": fmt.Sprintf("%d", listing.Rooms)},
		{"area": fmt.Sprintf("%.2f", listing.Area)},
		{"floor": fmt.Sprintf("%d/%d", listing.Floor, listing.Floors)},
		{"series": listing.Series},
	}

	tableBody := ""
	for _, row := range rows {
		for k, v := range row {
			tableBody += fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>\n", k, v)
		}
	}

	body := fmt.Sprintf("<table border=\"1\" cellpadding=\"10\" cellspacing=\"0\">%s</table>", tableBody)

	return e.send(email.To, "listing "+listing.Id, body)
}

func (e *EmailClient) send(to string, subject string, body string) error {
	from := "me"
	msg := gmail.Message{
		Raw: base64.StdEncoding.EncodeToString([]byte(
			"From: " + from + "\r\n" +
				"To: " + to + "\r\n" +
				"Subject: " + subject + "\r\n" +
				"Content-Type: text/html; charset=UTF-8\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n" +
				"\r\n" + body,
		)),
	}

	_, err := e.gmailSvc.Users.Messages.Send("me", &msg).Do()
	return err
}
