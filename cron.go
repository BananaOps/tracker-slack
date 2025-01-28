package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bananaops/tracker/generated/proto/event/v1alpha1"
	"github.com/slack-go/slack"
)


type TodayEventsResponse struct {
	v1alpha1.TodayEventsResponse
}


func listEventToday() {
	api := slack.New(botToken)

	events, err := fetchEvents()
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des événements : %v", err)
		return
	}

	message := formatSlackMessageByEnvironment(events)

	channelID, slackTimestamp, err := api.PostMessage(
		os.Getenv("TRACKER_SLACK_CHANNEL"),
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		fmt.Printf("Error posting message to thread: %v", err)
	}

	fmt.Printf("Cron Message successfully sent to channel %s at %s \n", channelID, slackTimestamp)
}

// fetchEvents récupère les événements du jour depuis l'API
func fetchEvents() ([]*v1alpha1.Event, error) {
	resp, err := http.Get(os.Getenv("TRACKER_HOST") + "/api/v1alpha1/event")
	if err != nil {
		return []*v1alpha1.Event{}, fmt.Errorf("erreur lors de l'appel API : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []*v1alpha1.Event{}, fmt.Errorf("erreur API : statut %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []*v1alpha1.Event{}, fmt.Errorf("erreur lors de la lecture de la réponse : %v", err)
	}

	var todayEvents TodayEventsResponse
	if err := json.Unmarshal(body, &todayEvents); err != nil {
		return []*v1alpha1.Event{}, fmt.Errorf("erreur lors du parsing JSON : %v", err)
	}

	return todayEvents.Events, nil
}

// formatSlackMessageByEnvironment génère un texte groupé par environnement
func formatSlackMessageByEnvironment(events []*v1alpha1.Event) string {
	if len(events) == 0 {
		return "Aucun événement trouvé pour aujourd'hui !"
	}

	// Regrouper les événements par environnement
	groupedEvents := make(map[string][]*v1alpha1.Event)
	for _, event := range events {
		groupedEvents[event.Attributes.Environment.String()] = append(groupedEvents[event.Attributes.Environment.String()], event)
	}

	// Construire le message Slack
	message := "*Événements du jour regroupés par environnement :*\n"
	for env, envEvents := range groupedEvents {
		message += fmt.Sprintf("\n*Environnement : %s*\n", env)
		for _, event := range envEvents {
			message += fmt.Sprintf("- *%s*: <@%s>\n", event.Title, event.Attributes.Owner)
		}
	}

	return message
}
