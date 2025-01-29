package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type TodayEventReponse struct {
	Attributes struct {
		Message      string   `json:"message"`
		Priority     string   `json:"priority"`
		Service      string   `json:"service"`
		Source       string   `json:"source"`
		Status       string   `json:"status"`
		Type         string   `json:"type"`
		Environment  string   `json:"environment"`
		Impact       bool     `json:"impact"`
		StartDate    string   `json:"startDate"`
		EndDate      string   `json:"endDate"`
		Owner        string   `json:"owner"`
		StackHolders []string `json:"stackHolders"`
		Notification bool     `json:"notification"`
	} `json:"attributes"`
	Links struct {
		PullRequestLink string `json:"pullRequestLink"`
		Ticket          string `json:"ticket"`
	} `json:"links"`
	Metadata struct {
		SlackId string `json:"slackId"`
	} `json:"metadata"`
	Title string `json:"title"`
}

type TodayReponse struct {
	Events     []TodayEventReponse `json:"events"`
	TotalCount int                 `json:"totalcount"`
}

var channel string = os.Getenv("TRACKER_SLACK_CHANNEL")
var workspace string = os.Getenv("TRACKER_SLACK_WORKSPACE")

func listEventToday() {
	api := slack.New(botToken)

	events, err := fetchEvents()
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des événements : %v\n", err)
		return
	}

	message := formatSlackMessageByEnvironment(events)

	channelID, slackTimestamp, err := api.PostMessage(
		os.Getenv("TRACKER_SLACK_CHANNEL"),
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(true),        // false = Active Markdown (mrkdwn)
		slack.MsgOptionDisableLinkUnfurl(), // Désactive la preview des liens
	)
	if err != nil {
		fmt.Printf("Error posting message to thread: %v", err)
	}

	fmt.Printf("Cron Message successfully sent to channel %s at %s \n", channelID, slackTimestamp)

}

// fetchEvents récupère les événements du jour depuis l'API
func fetchEvents() ([]TodayEventReponse, error) {
	resp, err := http.Get(os.Getenv("TRACKER_HOST") + "/api/v1alpha1/events/today")
	if err != nil {
		return []TodayEventReponse{}, fmt.Errorf("erreur lors de l'appel API : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []TodayEventReponse{}, fmt.Errorf("erreur API : statut %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []TodayEventReponse{}, fmt.Errorf("erreur lors de la lecture de la réponse : %v", err)
	}

	var todayEvents TodayReponse
	if err := json.Unmarshal(body, &todayEvents); err != nil {
		return []TodayEventReponse{}, fmt.Errorf("erreur lors du parsing JSON : %v", err)
	}

	return todayEvents.Events, nil
}

// formatSlackMessageByEnvironment génère un texte groupé par environnement
func formatSlackMessageByEnvironment(events []TodayEventReponse) string {
	if len(events) == 0 {
		return ":calendar: No Tracker Events Today :rocket:"
	}

	// Regrouper les événements par environnement et service
	groupedEvents := make(map[string]map[string][]TodayEventReponse) // Structure : {environnement: {projet: [événements]}}
	for _, event := range events {
		if groupedEvents[event.Attributes.Environment] == nil {
			groupedEvents[event.Attributes.Environment] = make(map[string][]TodayEventReponse)
		}
		groupedEvents[event.Attributes.Environment][event.Attributes.Service] = append(groupedEvents[event.Attributes.Environment][event.Attributes.Service], event)
	}

	// Construire le message Slack
	message := ":calendar: *Today Tracker Events :*\n"
	for env, services := range groupedEvents {
		emoji, envMessage := getEnvironmentEmoji(env)
		message += fmt.Sprintf("\n*%s %s*\n", emoji, envMessage)
		for service, serviceEvents := range services {
			message += fmt.Sprintf("  *%s:*\n", service)
			for _, event := range serviceEvents {
				t, err := time.Parse(time.RFC3339, event.Attributes.StartDate)
				if err != nil {
					fmt.Println("Parsing Error :", err)
				}
				messageURL := fmt.Sprintf("https://%s.slack.com/archives/%s/p%s",
					workspace, channel, strings.ReplaceAll(event.Metadata.SlackId, ".", ""))

				message += fmt.Sprintf("    -  %02d:%02d - %s <%s|thread>\n", t.Hour(), t.Minute(), event.Title, messageURL)
			}
		}
	}

	return message
}

// getEnvironmentEmoji retourne l'émoji correspondant à un environnement
func getEnvironmentEmoji(environment string) (string, string) {
	switch environment {
	case "production":
		return ":prod:", "PROD"
	case "preproduction":
		return ":prep:", "PREPROD"
	case "UAT":
		return ":uat:", "UAT"
	default:
		return ":question:", "UNKOWN"
	}
}
