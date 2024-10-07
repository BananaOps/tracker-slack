package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"time"

	"github.com/slack-go/slack"
)

var (
	botToken = os.Getenv("SLACK_BOT_TOKEN")
	//verificationToken = os.Getenv("SLACK_VERIFICATION_TOKEN")
	signingSecret = os.Getenv("SLACK_SIGNING_SECRET")
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
	Stackholders []string `json:"stackholders"`
	EndDate      int64    `json:"end_date"`
	ReleaseTeam  string   `json:"release_team"`
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
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	api := slack.New(botToken)
	modalRequest := generateModalRequest()
	_, err = api.OpenView(s.TriggerID, modalRequest)
	if err != nil {
		fmt.Printf("Error opening view: %s", err)
	}
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

	api := slack.New(botToken)

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
		tracker.Stackholders = values["stackholders"]["multi_users_select-action"].SelectedUsers
		tracker.Ticket = values["ticket"]["url_text_input-action"].Value
		tracker.PullRequest = values["pull_request"]["url_text_input-action"].Value
		tracker.Description = values["changelog"]["text_input-action"].Value
		tracker.Owner = i.User.Name
		tracker.ReleaseTeam = values["release"]["select_input-release"].SelectedOption.Value

		// Post tracker event
		go postTrackerEvent(tracker)

		blocks := blockMessage(tracker)
		_, _, err = api.PostMessage(os.Getenv("TRACKER_SLACK_CHANNEL"),
			slack.MsgOptionText("Tracker created", false),
			slack.MsgOptionBlocks(blocks...),
			slack.MsgOptionAsUser(false),
			slack.MsgOptionLinkNames(true),
		)
		// Wait for a few seconds to see result this code is necesarry due to slack server modal is going to be closed after the update
		time.Sleep(time.Second * 2)
		if err != nil {
			fmt.Printf("Error updating view: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case slack.InteractionTypeBlockActions:
		handleBlockActions(i)
	}

}

func handleBlockActions(callback slack.InteractionCallback) {
	for _, action := range callback.ActionCallback.BlockActions {
		switch action.ActionID {
		case "action-edit":

			var tracker tracker
			/*
				values := callback.View.State.Values
				tracker.Summary = values["summary"]["text_input-action"].Value
				tracker.Project = values["project"]["text_input-action"].Value
				tracker.Priority = values["priority"]["select_input-priority"].SelectedOption.Value
				tracker.Datetime = values["datetime"]["datetimepicker-action"].SelectedDateTime
				tracker.Stackholders = values["stackholders"]["multi_users_select-action"].SelectedUsers
				tracker.Ticket = values["ticket"]["url_text_input-action"].Value
				tracker.PullRequest = values["pull_request"]["url_text_input-action"].Value
				tracker.Description = values["description"]["text_input-action"].Value
			*/
			fmt.Println(callback.View)
			fmt.Println(callback.View.Hash)
			updateMessage(callback.Channel.ID, callback.Message.Timestamp, callback.View.Hash, callback.View.ID, tracker)

		case "action-approvers":
			postThreadApproval(callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)
		
		case "action-reject":
			postThreadReject(callback.Channel.ID, callback.Message.Timestamp, callback.User.Name)
		}

	}
}

func updateMessage(channelID string, messageTs string, viewHash string, viewId string, tracker tracker) {
	api := slack.New(botToken)

	modalRequest := generateModalRequest()
	_, err := api.UpdateView(modalRequest, "", viewHash, viewId)
	if err != nil {
		fmt.Printf("Error update view: %s\n", err)
	}

	blocks := blockMessage(tracker)
	_, _, _, err = api.UpdateMessage(
		channelID,
		messageTs,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		fmt.Printf("Error posting message to thread: %v", err)
	}
}

func postThreadApproval(channelID string, messageTs string, approver string) {
	api := slack.New(botToken)

	message := fmt.Sprintf("Approved by <@%s>", approver)
	_, _, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(messageTs),
	)
	if err != nil {
		fmt.Printf("Error posting message to thread: %v", err)
	}
}

func postThreadReject(channelID string, messageTs string, approver string) {
	api := slack.New(botToken)

	message := fmt.Sprintf("Rejected by <@%s>", approver)
	_, _, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(messageTs),
	)
	if err != nil {
		fmt.Printf("Error posting message to thread: %v", err)
	}
}

type Payload struct {
	Attributes struct {
		Message     string `json:"message"`
		Priority    int    `json:"priority"`
		Service     string `json:"service"`
		Source      string `json:"source"`
		Status      int    `json:"status"`
		Type        int    `json:"type"`
		Environment int    `json:"environment"`
		Impact      bool   `json:"impact"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Owner       string `json:"owner"`
	} `json:"attributes"`
	Links struct {
		PullRequestLink string `json:"pull_request_link"`
		Ticket          string `json:"ticket"`
	} `json:"links"`
	Title string `json:"title"`
}

var environment map[string]int = map[string]int{"PROD": 7, "PREP": 6, "UAT": 4}

func postTrackerEvent(tracker tracker) {

	var data Payload

	data.Attributes.Message = tracker.Summary
	data.Attributes.Priority = 1
	data.Attributes.Service = tracker.Project
	data.Attributes.Source = "slack"
	data.Attributes.Status = 1
	data.Attributes.Type = 1
	data.Attributes.Environment = environment[tracker.Environment]
	data.Attributes.Impact = true
	data.Attributes.StartDate = time.Unix(tracker.Datetime, 0).Format("2006-01-02T15:04:05Z")
	if tracker.EndDate == 0 {
		tracker.EndDate = tracker.Datetime + 3600
	}
	data.Attributes.EndDate = time.Unix(tracker.EndDate, 0).Format("2006-01-02T15:04:05Z")
	data.Attributes.Owner = tracker.Owner
	data.Links.PullRequestLink = tracker.PullRequest
	data.Links.Ticket = tracker.Ticket
	data.Title = tracker.Summary

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}

	//fmt.Println(string(payloadBytes))

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", os.Getenv("TRACKER_HOST")+"/api/v1alpha1/event", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
}
