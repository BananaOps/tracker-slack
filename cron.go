package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type TodayEventReponse struct {
	Attributes struct {
		Message       string   `json:"message"`
		Priority      string   `json:"priority"`
		Service       string   `json:"service"`
		Source        string   `json:"source"`
		Status        string   `json:"status"`
		Type          string   `json:"type"`
		Environment   string   `json:"environment"`
		Impact        bool     `json:"impact"`
		StartDate     string   `json:"startDate"`
		EndDate       string   `json:"endDate"`
		Owner         string   `json:"owner"`
		StakeHolders  []string `json:"stakeHolders"`
		Notification  bool     `json:"notification"`
		Notifications []string `json:"notifications"`
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

var channel string = os.Getenv("TRACKER_DEPLOYMENT_CHANNEL")
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
		os.Getenv("TRACKER_DEPLOYMENT_CHANNEL"),
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
				time, err := convertTimeLocation(event.Attributes.StartDate)
				if err != nil {
					fmt.Printf("Error to convert time with location: %v\n", err)
					continue

				}
				messageURL := createSlackMessageURL(workspace, channel, event.Metadata.SlackId)
				statusIcon := getStatusIcon(event.Attributes.Status)

				message += fmt.Sprintf("    - %s %s - %s %s\n", statusIcon, time, event.Title, messageURL)
			}
		}
	}

	return message
}

func convertTimeLocation(StartDate string) (string, error) {

	t, err := time.Parse(time.RFC3339, StartDate)
	if err != nil {
		return "", err
	}
	//To convert print datetime in location
	location, err := time.LoadLocation(os.Getenv("TRACKER_TIMEZONE"))
	if err != nil {
		return "", err
	}
	timeInUTCLocation := t.In(location)
	formattedTime := timeInUTCLocation.Format("15:04")

	return fmt.Sprintf("%s %s", formattedTime, location.String()), nil
}

// isValidSlackTimestamp vérifie si le slackId est un timestamp Slack valide
func isValidSlackTimestamp(slackId string) bool {
	if slackId == "" {
		return false
	}
	// Vérifie si le slackId correspond au format d'un timestamp Slack
	match, _ := regexp.MatchString(`^\d+\.\d+$`, slackId)
	return match
}

// createSlackMessageURL crée l'URL du message Slack si le slackId est valide
func createSlackMessageURL(teamDomain, channelId, slackId string) string {
	if !isValidSlackTimestamp(slackId) {
		return ""
	}
	return fmt.Sprintf("<https://%s.slack.com/archives/%s/p%s|thread>", teamDomain, channelId, strings.ReplaceAll(slackId, ".", ""))
}

// getEnvironmentEmoji retourne l'émoji correspondant à un environnement
func getEnvironmentEmoji(environment string) (string, string) {
	switch environment {
	case "production":
		return "🔴", "PROD"
	case "preproduction":
		return "🟡", "PREPROD"
	case "UAT":
		return "🔵", "UAT"
	case "development":
		return "🟢", "DEV"
	default:
		return "❓", "UNKOWN"
	}
}

// getStatusIcon retourne l'icône correspondant au statut d'un événement
func getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "failed", "fail":
		return "❌"
	case "cancelled", "canceled", "cancel":
		return "❌"
	case "done", "completed", "success":
		return "✅"
	case "in_progress", "inprogress", "running":
		return "🔄"
	case "paused", "pause":
		return "⏸️"
	case "pending", "scheduled":
		return "⏳"
	default:
		return "📋" // Icône par défaut pour les statuts inconnus
	}
}
