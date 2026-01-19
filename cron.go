package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
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
	case "failed", "fail", "failure":
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

// syncEventsToSlack synchronise les événements du tracker vers Slack
// Cette fonction récupère les événements du jour sans source Slack et sans slackId
// puis poste les messages Slack correspondants et met à jour le slackId
func syncEventsToSlack() {
	logger.Info("Starting sync events to Slack")

	// Récupérer les événements à synchroniser
	events, err := fetchEventsToSync()
	if err != nil {
		logger.Error("Failed to fetch events to sync", slog.Any("error", err))
		return
	}

	if len(events) == 0 {
		logger.Debug("No events to sync")
		return
	}

	logger.Info("Found events to sync", slog.Int("count", len(events)))

	api := slack.New(botToken)

	// Traiter chaque événement
	for _, event := range events {
		err := syncEventToSlack(api, event)
		if err != nil {
			logger.Error("Failed to sync event",
				slog.String("event_id", event.Metadata.Id),
				slog.String("title", event.Title),
				slog.Any("error", err))
			continue
		}
		logger.Info("Event synced successfully",
			slog.String("event_id", event.Metadata.Id),
			slog.String("title", event.Title))
	}

	logger.Info("Sync events to Slack completed")
}

// EventToSync représente un événement à synchroniser
type EventToSync struct {
	Attributes struct {
		Message       string   `json:"message"`
		Priority      string   `json:"priority"`
		Service       string   `json:"service"`
		Source        string   `json:"source"`
		Status        string   `json:"status"`
		Type          string   `json:"type"` // Type est une string dans l'API
		Environment   string   `json:"environment"`
		Impact        bool     `json:"impact"`
		StartDate     string   `json:"startDate"`
		EndDate       string   `json:"endDate"`
		Owner         string   `json:"owner"`
		StakeHolders  []string `json:"stakeHolders"`
		Notifications []string `json:"notifications"`
	} `json:"attributes"`
	Links struct {
		PullRequestLink string `json:"pullRequestLink"`
		Ticket          string `json:"ticket"`
	} `json:"links"`
	Metadata struct {
		Id      string `json:"id"`
		SlackId string `json:"slackId"`
	} `json:"metadata"`
	Title string `json:"title"`
}

// GetTypeAsInt convertit le type string en int
func (e *EventToSync) GetTypeAsInt() (int, error) {
	typeMap := map[string]int{
		"deployment": 1,
		"operation":  2,
		"drift":      3,
		"incident":   4,
		"rpa_usage":  5,
	}

	if typeInt, ok := typeMap[strings.ToLower(e.Attributes.Type)]; ok {
		return typeInt, nil
	}

	// Si c'est déjà un nombre en string, le convertir
	if typeInt, err := strconv.Atoi(e.Attributes.Type); err == nil {
		return typeInt, nil
	}

	return 0, fmt.Errorf("unknown event type: %s", e.Attributes.Type)
}

type EventsToSyncResponse struct {
	Events     []EventToSync `json:"events"`
	TotalCount int           `json:"totalcount"`
}

// fetchEventsToSync récupère les événements du jour sans source Slack et sans slackId
func fetchEventsToSync() ([]EventToSync, error) {
	// Calculer les dates de début et fin du jour (00:00:00 à 23:59:59)
	location, err := time.LoadLocation(os.Getenv("TRACKER_TIMEZONE"))
	if err != nil {
		location = time.UTC
		logger.Warn("Failed to load timezone, using UTC", slog.Any("error", err))
	}

	now := time.Now().In(location)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, location)

	// Formater les dates en RFC3339 et encoder pour l'URL
	startDate := startOfDay.Format(time.RFC3339)
	endDate := endOfDay.Format(time.RFC3339)

	// Construire l'URL avec les paramètres de recherche encodés
	baseURL := os.Getenv("TRACKER_HOST") + "/api/v1alpha1/events/search"
	params := url.Values{}
	params.Add("start_date", startDate)
	params.Add("end_date", endDate)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	logger.Debug("Fetching events to sync",
		slog.String("start_date", startDate),
		slog.String("end_date", endDate),
		slog.String("url", fullURL))

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response EventsToSyncResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Filtrer les événements : source != "slack" ET slackId vide
	var filteredEvents []EventToSync
	for _, event := range response.Events {
		if event.Attributes.Source != "slack" && event.Metadata.SlackId == "" {
			filteredEvents = append(filteredEvents, event)
		}
	}

	logger.Debug("Events filtered for sync",
		slog.Int("total", len(response.Events)),
		slog.Int("to_sync", len(filteredEvents)))

	return filteredEvents, nil
}

// syncEventToSlack poste un événement sur Slack et met à jour le slackId
func syncEventToSlack(api *slack.Client, event EventToSync) error {
	// Convertir le type string en int
	eventType, err := event.GetTypeAsInt()
	if err != nil {
		return fmt.Errorf("failed to get event type: %w", err)
	}

	// Convertir l'événement en tracker pour utiliser les fonctions existantes
	tracker := convertEventToTracker(event)

	// Déterminer le channel et les blocks selon le type
	var channelID string
	var blocks []slack.Block

	switch eventType {
	case 1: // Deployment
		channelID = os.Getenv("TRACKER_DEPLOYMENT_CHANNEL")
		blocks = blockDeploymentMessage(tracker)
	case 2: // Operation
		channelID = os.Getenv("TRACKER_OPERATION_CHANNEL")
		blocks = blockOperationMessage(tracker)
	case 3: // Drift
		channelID = os.Getenv("TRACKER_DRIFT_CHANNEL")
		blocks = blockDriftMessage(tracker)
	case 4: // Incident
		channelID = os.Getenv("TRACKER_INCIDENT_CHANNEL")
		blocks = blockIncidentMessage(tracker)
	case 5: // RPA Usage
		channelID = os.Getenv("TRACKER_RPA_USAGE_CHANNEL")
		blocks = blockRPAUsageMessage(tracker)
	default:
		return fmt.Errorf("unknown event type: %d", eventType)
	}

	// Poster le message sur Slack
	_, slackTimestamp, err := api.PostMessage(channelID,
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	logger.Info("Message posted to Slack",
		slog.String("channel", channelID),
		slog.String("timestamp", slackTimestamp))

	// Mettre à jour le slackId dans le tracker
	err = updateTrackerEventSlackId(event.Metadata.Id, slackTimestamp)
	if err != nil {
		return fmt.Errorf("failed to update slackId: %w", err)
	}

	return nil
}

// convertEventToTracker convertit un EventToSync en tracker
func convertEventToTracker(event EventToSync) tracker {
	// Parser les dates
	startDate, _ := time.Parse(time.RFC3339, event.Attributes.StartDate)
	endDate, _ := time.Parse(time.RFC3339, event.Attributes.EndDate)

	// Convertir le type string en int
	eventType, err := event.GetTypeAsInt()
	if err != nil {
		logger.Warn("Failed to convert event type, using default",
			slog.String("type", event.Attributes.Type),
			slog.Any("error", err))
		eventType = 1 // Default to deployment
	}

	// Convertir impact bool en string
	impact := "No"
	if event.Attributes.Impact {
		impact = "Yes"
	}

	// Convertir notifications en release/support team
	releaseTeam := "No"
	supportTeam := "No"
	for _, notif := range event.Attributes.Notifications {
		if strings.EqualFold(notif, "release") {
			releaseTeam = "Yes"
		}
		if strings.EqualFold(notif, "support") {
			supportTeam = "Yes"
		}
	}

	return tracker{
		Type:         eventType,
		Datetime:     startDate.Unix(),
		Summary:      event.Title,
		Project:      event.Attributes.Service,
		Priority:     event.Attributes.Priority,
		Environment:  mapEnvironmentToCode(event.Attributes.Environment),
		Impact:       impact,
		Ticket:       event.Links.Ticket,
		PullRequest:  event.Links.PullRequestLink,
		Description:  event.Attributes.Message,
		Owner:        event.Attributes.Owner,
		Stakeholders: event.Attributes.StakeHolders,
		EndDate:      endDate.Unix(),
		ReleaseTeam:  releaseTeam,
		SupportTeam:  supportTeam,
		SlackId:      event.Metadata.SlackId,
	}
}

// mapEnvironmentToCode convertit le nom d'environnement en code
func mapEnvironmentToCode(env string) string {
	switch env {
	case "production":
		return "PROD"
	case "preproduction":
		return "PREP"
	case "UAT":
		return "UAT"
	case "development":
		return "DEV"
	default:
		return "PROD"
	}
}
