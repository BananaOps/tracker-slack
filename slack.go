package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"time"

	"github.com/slack-go/slack"
)

var (
	botToken = os.Getenv("SLACK_BOT_TOKEN")
	//verificationToken = os.Getenv("SLACK_VERIFICATION_TOKEN")
	signingSecret    = os.Getenv("SLACK_SIGNING_SECRET")
	messageTimestamp string
	messageChannel   string
)

type tracker struct {
	Datetime     int64    `json:"datetime"`
	Summary      string   `json:"summary"`
	Project      string   `json:"project"`
	Environment  string   `json:"environment"`
	Impact       string   `json:"impact"`
	Ticket       string   `json:"ticket"`
	PullRequest  string   `json:"pull_request"`
	Description  string   `json:"description"`
	Owner        string   `json:"owner"`
	Stakeholders []string `json:"stakeHolders"`
	EndDate      int64    `json:"end_date"`
	ReleaseTeam  string   `json:"release_team"`
	SupportTeam  string   `json:"support_team"`
	SlackId      string   `json:"slack_id"`
}

// This was taken from the slash example
// https://github.com/slack-go/slack/blob/master/examples/slash/slash.go
func verifySigningSecret(r *http.Request) error {
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	// Need to use r.Body again when unmarshalling SlashCommand and InteractionCallback
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	_, err = verifier.Write(body)
	if err != nil {
		fmt.Println(err.Error())
	}

	if err = verifier.Ensure(); err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

// handleCommand processes incoming Slack commands
func handleCommand(w http.ResponseWriter, r *http.Request) {

	// check if the request is authorized
	err := verifySigningSecret(r)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		http.Error(w, "Failed to parse Slack command", http.StatusInternalServerError)
		return
	}

	switch s.Command {
	case "/deployment":
		handleDeploymentCommand(w, s)
	case "/incident":
		handleIncidentCommand(w, s)
	case "/drift":
		handleDriftCommand(w, s)
	default:
		http.Error(w, "Unknown command", http.StatusBadRequest)
	}
}

// handleDeploymentCommand handles the /deployment command
func handleDeploymentCommand(w http.ResponseWriter, s slack.SlashCommand) {
	api := slack.New(botToken)
	view := generateModalRequest(EventReponse{})
	_, err := api.OpenView(s.TriggerID, view)
	if err != nil {
		fmt.Printf("Error opening view: %s", err)
	}
	w.WriteHeader(http.StatusOK)
}

// handleIncidentCommand handles the /incident command
func handleIncidentCommand(w http.ResponseWriter, s slack.SlashCommand) {
	response := fmt.Sprintf("Handling incident command with text: %s", s.Text)
	w.Write([]byte(response))
}

// handleDriftCommand handles the /drift command
func handleDriftCommand(w http.ResponseWriter, s slack.SlashCommand) {
	response := fmt.Sprintf("Handling drift command with text: %s", s.Text)
	w.Write([]byte(response))
}


func handleInteractiveAPIEndpoint(w http.ResponseWriter, r *http.Request) {

	// check if the request is authorized
	err := verifySigningSecret(r)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var i slack.InteractionCallback
	err = json.Unmarshal([]byte(r.FormValue("payload")), &i)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// update modal sample
	switch i.Type {
	//update when interaction type is view_submission
	case slack.InteractionTypeViewSubmission:

		var tracker tracker
		values := i.View.State.Values

		tracker.Summary = values["summary"]["text_input-action"].Value
		tracker.Project = values["project"]["text_input-action"].Value
		tracker.Environment = values["environment"]["select_input-environment"].SelectedOption.Value
		tracker.Impact = values["impact"]["select_input-impact"].SelectedOption.Value
		tracker.Datetime = values["datetime"]["datetimepicker-action"].SelectedDateTime
		tracker.EndDate = values["enddatetime"]["datetimepicker-action"].SelectedDateTime
		tracker.Stakeholders = values["stakeholders"]["multi_users_select-action"].SelectedUsers
		tracker.Ticket = values["ticket"]["url_text_input-action"].Value
		tracker.PullRequest = values["pull_request"]["url_text_input-action"].Value
		tracker.Description = values["changelog"]["text_input-action"].Value
		tracker.Owner = i.User.Name
		tracker.ReleaseTeam = values["release"]["select_input-release"].SelectedOption.Value
		tracker.SupportTeam = values["support"]["select_input-support"].SelectedOption.Value

		api := slack.New(botToken)
		var channelID string
		var slackTimestamp string

		if i.View.CallbackID == "edit" {

			event := getTrackerEvent(messageTimestamp)
			tracker.Owner = event.Event.Attributes.Owner

			blocks := blockMessage(tracker)

			channelID, slackTimestamp, _, err := api.UpdateMessage(messageChannel,
				messageTimestamp,
				slack.MsgOptionBlocks(blocks...),
			)
			if err != nil {
				fmt.Printf("Error updating message: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			_, _, err = api.PostMessage(
				messageChannel,
				slack.MsgOptionText(fmt.Sprintf(":pencil: Edited by <@%s>", i.User.ID), false),
				slack.MsgOptionTS(messageTimestamp),
			)
			if err != nil {
				fmt.Printf("Error posting message to thread: %v", err)
			}
			fmt.Printf("Message successfully updated to channel %s at %s \n", channelID, slackTimestamp)

			// Post tracker event
			tracker.SlackId = string(messageTimestamp)
			go editTrackerEvent(tracker)

		} else {

			blocks := blockMessage(tracker)

			channelID, slackTimestamp, err = api.PostMessage(os.Getenv("TRACKER_SLACK_CHANNEL"),
				slack.MsgOptionBlocks(blocks...),
			)
			if err != nil {
				fmt.Printf("Error posting message: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			fmt.Printf("Message successfully sent to channel %s at %s \n", channelID, slackTimestamp)
			// Post tracker event
			tracker.SlackId = string(slackTimestamp)
			go postTrackerEvent(tracker)
		}

		// Wait for a few seconds to see result this code is necesarry due to slack server modal is going to be closed after the update
		time.Sleep(time.Second * 2)

	case slack.InteractionTypeBlockActions:
		handleBlockActions(i, w)
	}
}

func handleBlockActions(callback slack.InteractionCallback, w http.ResponseWriter) {
	for _, action := range callback.ActionCallback.BlockActions {

		switch action.ActionID {
		case "action-edit":
			messageTimestamp = callback.Message.Timestamp
			messageChannel = callback.Channel.ID
			event := getTrackerEvent(messageTimestamp)
			api := slack.New(botToken)
			view := generateModalRequest(event.Event)
			view.CallbackID = "edit"
			_, err := api.OpenView(callback.TriggerID, view)
			if err != nil {
				fmt.Printf("Error Open view: %s", err)
			}
			w.WriteHeader(http.StatusOK)

		case "action-approvers":
			postThreadAction("approved", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

		case "action-reject":
			postThreadAction("rejected", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

		case "action":
			switch action.SelectedOption.Value {
			case "in_progress":
				postThreadAction("in_progress", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

			case "pause":
				postThreadAction("pause", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

			case "cancelled":
				event := getTrackerEvent(callback.Message.Timestamp)
				updateTrackerEvent(event.Event, 2)
				postThreadAction("cancelled", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

			case "post_poned":
				postThreadAction("post_poned", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

			case "done":
				event := getTrackerEvent(callback.Message.Timestamp)
				fmt.Println(event)
				updateTrackerEvent(event.Event, 3)
				postThreadAction("done", callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)

			}
		}

	}
}

func postThreadAction(action string, channelID string, messageTs string, user string) {
	api := slack.New(botToken)

	var message string
	var reaction string
	switch action {
	case "in_progress":
		message = fmt.Sprintf(":loading: In progress by <@%s>", user)
		reaction = "loading"
	case "pause":
		message = fmt.Sprintf(":double_vertical_bar: Paused by <@%s>", user)
		reaction = "double_vertical_bar"
	case "cancelled":
		message = fmt.Sprintf(":x: Cancelled by <@%s>", user)
		reaction = "x"
	case "post_poned":
		message = fmt.Sprintf(":hourglass_flowing_sand: Postponed by <@%s>", user)
		reaction = "hourglass_flowing_sand"
	case "done":
		message = fmt.Sprintf(":white_check_mark: Done by <@%s>", user)
		reaction = "white_check_mark"
	case "approved":
		message = fmt.Sprintf(":ok: Approved by <@%s>", user)
		reaction = "ok"
	case "rejected":
		message = fmt.Sprintf(":x: Rejected by <@%s>", user)
		reaction = "x"
	case "edit":
		message = fmt.Sprintf(":pencil: Edited by <@%s>", user)
		reaction = "pencil"
	}

	_, _, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(messageTs),
	)
	if err != nil {
		fmt.Printf("Error posting message to thread: %v", err)
	}

	err = messageReaction(api, channelID, messageTs, reaction)
	if err != nil {
		fmt.Printf("Error manage reaction: %v", err)
	}
}

func messageReaction(api *slack.Client, channelID string, messageTs string, reaction string) error {

	itemRef := slack.ItemRef{
		Channel:   channelID,
		Timestamp: messageTs,
	}

	trackerReactions, err := api.GetReactions(
		itemRef,
		slack.GetReactionsParameters{
			Full: true,
		})

	if err != nil {
		return err
	}

	for _, reaction := range trackerReactions {
		err = api.RemoveReaction(reaction.Name, itemRef)
		if err != nil {
			return err
		}
	}

	err = api.AddReaction(reaction, itemRef)
	if err != nil {
		return err
	}
	return nil
}

type Payload struct {
	Attributes struct {
		Message       string   `json:"message"`
		Priority      int      `json:"priority"`
		Service       string   `json:"service"`
		Source        string   `json:"source"`
		Status        int      `json:"status"`
		Type          int      `json:"type"`
		Environment   int      `json:"environment"`
		Impact        bool     `json:"impact"`
		StartDate     string   `json:"start_date"`
		EndDate       string   `json:"end_date"`
		Owner         string   `json:"owner"`
		StakeHolders  []string `json:"stakeHolders"`
		Notification  bool     `json:"notification"`
		Notifications []string `json:"notifications"`
	} `json:"attributes"`
	Links struct {
		PullRequestLink string `json:"pull_request_link"`
		Ticket          string `json:"ticket"`
	} `json:"links"`
	Title   string `json:"title"`
	SlackId string `json:"slack_id"`
}

type EventReponse struct {
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
		SlackId  string `json:"slackId"`
		Duration string `json:"duration"`
		Id       string `json:"id"`
	}
	Title string `json:"title"`
}

type Response struct {
	Event EventReponse `json:"event"`
}

var environment map[string]int = map[string]int{"PROD": 7, "PREP": 6, "UAT": 4, "DEV": 1}

func postTrackerEvent(tracker tracker) {

	var data Payload

	data.Attributes.Message = tracker.Description
	data.Attributes.Priority = 1
	data.Attributes.Service = tracker.Project
	data.Attributes.Source = "slack"
	data.Attributes.Status = 1
	data.Attributes.Type = 1
	data.Attributes.Environment = environment[tracker.Environment]
	if tracker.Impact == "Yes" {
		data.Attributes.Impact = true
	} else {
		data.Attributes.Impact = false
	}
	data.Attributes.StartDate = time.Unix(tracker.Datetime, 0).Format("2006-01-02T15:04:05Z")
	if tracker.EndDate == 0 {
		tracker.EndDate = tracker.Datetime + 3600
	}
	data.Attributes.EndDate = time.Unix(tracker.EndDate, 0).Format("2006-01-02T15:04:05Z")
	data.Attributes.Owner = tracker.Owner
	data.Links.PullRequestLink = tracker.PullRequest
	data.Links.Ticket = tracker.Ticket
	data.Attributes.StakeHolders = tracker.Stakeholders
	if tracker.ReleaseTeam == "Yes" {
		data.Attributes.Notifications = append(data.Attributes.Notifications, "release")
	}
	if tracker.SupportTeam == "Yes" {
		data.Attributes.Notifications = append(data.Attributes.Notifications, "support")
	}
	if IsValidURL(tracker.Ticket) || tracker.Ticket == "" {
		data.Links.PullRequestLink = tracker.PullRequest
	} else {
		fmt.Printf("Invalid PullRequest URL: %s\n", tracker.PullRequest)
	}
	if IsValidURL(tracker.Ticket) || tracker.Ticket == "" {
		data.Links.Ticket = tracker.Ticket
	} else {
		fmt.Printf("Invalid Ticket URL: %s\n", tracker.Ticket)
	}
	data.Title = tracker.Summary
	data.SlackId = tracker.SlackId

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", os.Getenv("TRACKER_HOST")+"/api/v1alpha1/event", body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}

func editTrackerEvent(tracker tracker) {

	var data Payload

	data.Attributes.Message = tracker.Description
	data.Attributes.Priority = 1
	data.Attributes.Service = tracker.Project
	data.Attributes.Source = "slack"
	data.Attributes.Status = 7
	data.Attributes.Type = 1
	data.Attributes.Environment = environment[tracker.Environment]
	if tracker.Impact == "Yes" {
		data.Attributes.Impact = true
	} else {
		data.Attributes.Impact = false
	}
	data.Attributes.StartDate = time.Unix(tracker.Datetime, 0).Format("2006-01-02T15:04:05Z")
	if tracker.EndDate == 0 {
		tracker.EndDate = tracker.Datetime + 3600
	}
	data.Attributes.EndDate = time.Unix(tracker.EndDate, 0).Format("2006-01-02T15:04:05Z")
	data.Attributes.Owner = tracker.Owner
	if IsValidURL(tracker.Ticket) || tracker.Ticket == "" {
		data.Links.PullRequestLink = tracker.PullRequest
	} else {
		fmt.Printf("Invalid PullRequest URL: %s\n", tracker.PullRequest)
	}
	if IsValidURL(tracker.Ticket) || tracker.Ticket == "" {
		data.Links.Ticket = tracker.Ticket
	} else {
		fmt.Printf("Invalid Ticket URL: %s\n", tracker.Ticket)
	}
	data.Title = tracker.Summary
	data.SlackId = tracker.SlackId
	data.Attributes.StakeHolders = tracker.Stakeholders
	if tracker.ReleaseTeam == "Yes" {
		data.Attributes.Notifications = append(data.Attributes.Notifications, "release")
	}
	if tracker.SupportTeam == "Yes" {
		data.Attributes.Notifications = append(data.Attributes.Notifications, "support")
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("PUT", os.Getenv("TRACKER_HOST")+"/api/v1alpha1/event", body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}

var environmentMap map[string]int = map[string]int{"production": 7, "preproduction": 6, "UAT": 4, "development": 1}

func updateTrackerEvent(tracker EventReponse, status int) {

	var data Payload

	data.Attributes.Message = tracker.Attributes.Message
	data.Attributes.Priority = 1
	data.Attributes.Service = tracker.Attributes.Service
	data.Attributes.Source = "slack"
	data.Attributes.Status = status
	data.Attributes.Type = 1
	fmt.Println(tracker)
	data.Attributes.Environment = environmentMap[tracker.Attributes.Environment]
	data.Attributes.Impact = tracker.Attributes.Impact
	data.Attributes.StartDate = tracker.Attributes.StartDate
	data.Attributes.EndDate = tracker.Attributes.EndDate
	data.Attributes.Owner = tracker.Attributes.Owner
	data.Links.PullRequestLink = tracker.Links.PullRequestLink
	data.Links.Ticket = tracker.Links.Ticket
	data.Title = tracker.Title
	data.SlackId = tracker.Metadata.SlackId
	data.Attributes.StakeHolders = tracker.Attributes.StakeHolders
	data.Attributes.Notifications = tracker.Attributes.Notifications

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("PUT", os.Getenv("TRACKER_HOST")+"/api/v1alpha1/event", body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}

func getTrackerEvent(id string) Response {

	resp, err := http.Get(os.Getenv("TRACKER_HOST") + "/api/v1alpha1/event/" + id)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du corps : %s", err)
	}

	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du corps : %s", err)
	}
	return data
}

// Fonction pour valider si une string est une URL valide
func IsValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	// Vérifier si le schéma et l'hôte sont présents (http, https, etc.)
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
