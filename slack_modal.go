package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

func generateModalRequest(event EventReponse) slack.ModalViewRequest {

	pullRequest := inputUrl("pull_request", "Link Pull Request", event.Links.PullRequestLink, ":github:")
	pullRequest.Optional = true

	ticket := inputUrl("ticket", "Link Ticket Issue", event.Links.Ticket, ":ticket:")
	ticket.Optional = true

	stakeholders := inputMultiUser("stakeholders", ":dart: Stakeholders", event.Attributes.StakeHolders)
	stakeholders.Optional = true

	changelog := inputText("changelog", "Description", event.Attributes.Message, "", true)
	changelog.Optional = true

	endDateTime := inputDatetime("enddatetime", "End Date", event.Attributes.EndDate)
	endDateTime.Optional = true

	checkNotificationRelease := checkNotification(event.Attributes.Notifications, "release")
	checkNotificationSupport := checkNotification(event.Attributes.Notifications, "support")

	modalRequest := slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: "modal-identifier",
		Title:      slack.NewTextBlockObject("plain_text", "Tracker", true, false),
		Submit:     slack.NewTextBlockObject("plain_text", "Submit", true, false),
		Close:      slack.NewTextBlockObject("plain_text", "Cancel", true, false),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				inputText("summary", "Summary", event.Title, "", false),
				inputText("project", "Project", event.Attributes.Service, ":rocket:", false),
				inputEnv(event.Attributes.Environment),
				inputImpact(event.Attributes.Impact),
				inputReleaseTeam(checkNotificationRelease),
				inputSupportTeam(checkNotificationSupport),
				//inputAction(),
				inputDatetime("datetime", "Start Date", event.Attributes.StartDate),
				endDateTime,
				stakeholders,
				ticket,
				pullRequest,
				changelog,
			},
		},
	}

	return modalRequest
}

func checkNotification(notification []string, name string) bool {
	for i := range notification {
		if strings.EqualFold(notification[i], name) {
			return true
		}
	}
	return false
}

func blockMessage(tracker tracker) []slack.Block {

	var users []string

	for i := range tracker.Stakeholders {
		user := fmt.Sprintf("<@%s>", tracker.Stakeholders[i])
		users = append(users, user)
	}

	//var priorityEmoji map[string]string = map[string]string{"P1": ":priority-highest:", "P2": ":priority-high:", "P3": ":priority-medium:", "P4": ":priority-low:"}

	var priorityEnv map[string]string = map[string]string{"PROD": ":prod:", "PREP": ":prep:", "UAT": ":uat:", "DEV": ":development:"}

	//To convert print datetime in location
	t := time.Unix(tracker.Datetime, 0).UTC()
	location, err := time.LoadLocation(os.Getenv("TRACKER_TIMEZONE"))
	if err != nil {
		fmt.Println(err)
	}
	timeInUTCLocation := t.In(location)
	formattedTime := timeInUTCLocation.Format("2006-01-02 15:04")

	summary := fmt.Sprintf("*%s* \n \n", tracker.Summary)
	project := fmt.Sprintf(":rocket: *Project:* %s \n", tracker.Project)
	date := fmt.Sprintf(":date: *Date:* %s %s \n", formattedTime, location.String())
	environment := fmt.Sprintf("%s *Environment:* %s \n", priorityEnv[tracker.Environment], tracker.Environment)
	impact := fmt.Sprintf(":boom: *Impact:* %s \n", tracker.Impact)
	releaseTeam := ":slack_notification: *Notification Release:* @release-team \n"
	supportTeam := ":slack_notification: *Notification Support:* @team-support \n"
	owner := fmt.Sprintf(":technologist: *Owner:* <@%s> \n", tracker.Owner)
	description := fmt.Sprintf(":memo: *Description:* \n %s \n", tracker.Description)

	var stackholder string
	if len(users) > 0 {
		stackholder = fmt.Sprintf(":dart: *Stakeholders:* %s \n", strings.Join(users, ", "))
	}

	var pullRequest string
	if tracker.PullRequest != "" {
		pullRequest = fmt.Sprintf(":github: *Pull Request:* %s \n", tracker.PullRequest)
	}

	var ticket string
	if tracker.Ticket != "" {
		ticket = fmt.Sprintf(":ticket: *Ticket Issue:* %s \n", tracker.Ticket)
	}

	if tracker.ReleaseTeam == "No" {
		releaseTeam = ""
	}
	if tracker.SupportTeam == "No" {
		supportTeam = ""
	}

	message := summary + project + date + environment + impact + owner + releaseTeam + supportTeam + stackholder + ticket + pullRequest + description

	// Define the modal blocks
	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", message, false, false),
			nil,
			nil,
		),
		slack.NewActionBlock(
			"actionblock",
			slack.NewButtonBlockElement(
				"action-edit",
				"click_me_123",
				slack.NewTextBlockObject("plain_text", ":pencil: Edit", true, false),
			),
			slack.NewButtonBlockElement(
				"action-approvers",
				"click_me_123",
				slack.NewTextBlockObject("plain_text", ":ok: Approval", true, false),
			),
			slack.NewButtonBlockElement(
				"action-reject",
				"click_me_123",
				slack.NewTextBlockObject("plain_text", ":x: Reject", true, false),
			),
		),
		slack.NewActionBlock(
			"status",
			slack.NewOptionsSelectBlockElement(
				"static_select",
				slack.NewTextBlockObject("plain_text", "Select status", true, false),
				"action",
				//slack.NewOptionBlockObject("edit", slack.NewTextBlockObject("plain_text", ":note: Edit", true, false), nil),
				slack.NewOptionBlockObject("in_progress", slack.NewTextBlockObject("plain_text", ":loading: InProgress", true, false), nil),
				slack.NewOptionBlockObject("pause", slack.NewTextBlockObject("plain_text", ":double_vertical_bar: Pause", true, false), nil),
				slack.NewOptionBlockObject("cancelled", slack.NewTextBlockObject("plain_text", ":x: Cancelled", true, false), nil),
				slack.NewOptionBlockObject("post_poned", slack.NewTextBlockObject("plain_text", ":hourglass_flowing_sand: PostPoned", true, false), nil),
				slack.NewOptionBlockObject("done", slack.NewTextBlockObject("plain_text", ":white_check_mark: Done", true, false), nil),
			),
		),
	}

	return blocks
}

func inputMultiUser(blockId string, BlockText string, values []string) *slack.InputBlock {

	block := slack.NewOptionsMultiSelectBlockElement(
		slack.MultiOptTypeUser,
		slack.NewTextBlockObject("plain_text", "Select users", false, false),
		"multi_users_select-action",
	)
	block.InitialUsers = values

	return slack.NewInputBlock(
		blockId,
		slack.NewTextBlockObject("plain_text", BlockText, false, false),
		nil,
		block,
	)
}

func inputUrl(blockId string, blockText string, value string, emoji string) *slack.InputBlock {

	block := slack.NewURLTextInputBlockElement(slack.NewTextBlockObject("plain_text", blockText, true, false), "url_text_input-action")
	block.InitialValue = value

	return slack.NewInputBlock(
		blockId,
		slack.NewTextBlockObject("plain_text", fmt.Sprintf("%s %s", emoji, blockText), true, false),
		nil,
		block,
	)
}

func inputText(blockId string, blockText string, value string, emoji string, multiline bool) *slack.InputBlock {

	var InputBlockElement *slack.PlainTextInputBlockElement
	if multiline {
		InputBlockElement = slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", blockText, true, false), "text_input-action")
		InputBlockElement.Multiline = true
	} else {
		InputBlockElement = slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", blockText, true, false), "text_input-action")
	}
	InputBlockElement.InitialValue = value

	return slack.NewInputBlock(
		blockId,
		slack.NewTextBlockObject("plain_text", fmt.Sprintf("%s %s", emoji, blockText), true, false),
		nil,
		InputBlockElement,
	)
}

func inputImpact(value bool) *slack.InputBlock {

	block := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject("plain_text", "Change have an impact ?", true, false),
		"select_input-impact",
		slack.NewOptionBlockObject("Yes", slack.NewTextBlockObject("plain_text", "Yes", true, false), nil),
		slack.NewOptionBlockObject("No", slack.NewTextBlockObject("plain_text", "No", true, false), nil),
	)
	if value {
		block.InitialOption = slack.NewOptionBlockObject("Yes", slack.NewTextBlockObject("plain_text", "Yes", true, false), nil)
	} else {
		block.InitialOption = slack.NewOptionBlockObject("No", slack.NewTextBlockObject("plain_text", "No", true, false), nil)
	}

	return slack.NewInputBlock(
		"impact",
		slack.NewTextBlockObject("plain_text", ":boom: Impact", true, false),
		nil,
		block,
	)
}

func inputReleaseTeam(value bool) *slack.InputBlock {

	block := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject("plain_text", "Need notify Release Team ?", true, false),
		"select_input-release",
		slack.NewOptionBlockObject("Yes", slack.NewTextBlockObject("plain_text", "Yes", true, false), nil),
		slack.NewOptionBlockObject("No", slack.NewTextBlockObject("plain_text", "No", true, false), nil),
	)

	if value {
		block.InitialOption = slack.NewOptionBlockObject("Yes", slack.NewTextBlockObject("plain_text", "Yes", true, false), nil)
	} else {
		block.InitialOption = slack.NewOptionBlockObject("No", slack.NewTextBlockObject("plain_text", "No", true, false), nil)
	}

	return slack.NewInputBlock(
		"release",
		slack.NewTextBlockObject("plain_text", ":question: Notify Release Team ", true, false),
		nil,
		block,
	)
}

func inputSupportTeam(value bool) *slack.InputBlock {

	block := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject("plain_text", "Need notify Support Team ?", true, false),
		"select_input-support",
		slack.NewOptionBlockObject("Yes", slack.NewTextBlockObject("plain_text", "Yes", true, false), nil),
		slack.NewOptionBlockObject("No", slack.NewTextBlockObject("plain_text", "No", true, false), nil),
	)

	if value {
		block.InitialOption = slack.NewOptionBlockObject("Yes", slack.NewTextBlockObject("plain_text", "Yes", true, false), nil)
	} else {
		block.InitialOption = slack.NewOptionBlockObject("No", slack.NewTextBlockObject("plain_text", "No", true, false), nil)
	}

	return slack.NewInputBlock(
		"support",
		slack.NewTextBlockObject("plain_text", ":question: Notify Support Team ", true, false),
		nil,
		block,
	)
}

func inputEnv(value string) *slack.InputBlock {

	block := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject("plain_text", "Select Environment", true, false),
		"select_input-environment",
		slack.NewOptionBlockObject("PROD", slack.NewTextBlockObject("plain_text", "PROD", true, false), nil),
		slack.NewOptionBlockObject("PREP", slack.NewTextBlockObject("plain_text", "PREP", true, false), nil),
		slack.NewOptionBlockObject("UAT", slack.NewTextBlockObject("plain_text", "UAT", true, false), nil),
		slack.NewOptionBlockObject("DEV", slack.NewTextBlockObject("plain_text", "DEV", true, false), nil),
	)

	if value == "production" {
		block.InitialOption = slack.NewOptionBlockObject("PROD", slack.NewTextBlockObject("plain_text", "PROD", true, false), nil)
	} else if value == "preproduction" {
		block.InitialOption = slack.NewOptionBlockObject("PREP", slack.NewTextBlockObject("plain_text", "PREP", true, false), nil)
	} else if value == "UAT" {
		block.InitialOption = slack.NewOptionBlockObject("UAT", slack.NewTextBlockObject("plain_text", "UAT", true, false), nil)
	} else if value == "development" {
		block.InitialOption = slack.NewOptionBlockObject("DEV", slack.NewTextBlockObject("plain_text", "DEV", true, false), nil)
	} else {
		block.InitialOption = slack.NewOptionBlockObject("PROD", slack.NewTextBlockObject("plain_text", "PROD", true, false), nil)
	}

	return slack.NewInputBlock(
		"environment",
		slack.NewTextBlockObject("plain_text", ":prod: Environment", true, false),
		nil,
		block,
	)
}

func inputDatetime(blockId string, blockText string, value string) *slack.InputBlock {
	block := slack.NewDateTimePickerBlockElement("datetimepicker-action")
	if value == "" && blockId == "datetime" {
		block.InitialDateTime = time.Now().Unix()
	} else if value == "" && blockId == "enddatetime" {
		block.InitialDateTime = time.Now().Add(time.Hour * 1).Unix()
	} else {
		layout := time.RFC3339
		t, err := time.Parse(layout, value)
		if err != nil {
			fmt.Println("Erreur lors du parsing de la date :", err)
		}

		// Convertir la date en timestamp Unix (int64)
		timestamp := t.Unix()

		eventDatetime := timestamp
		if err != nil {
			fmt.Println(err)
		}
		block.InitialDateTime = eventDatetime
	}
	return slack.NewInputBlock(
		blockId,
		slack.NewTextBlockObject("plain_text", fmt.Sprintf(":date: %s", blockText), false, false),
		nil,
		block,
	)
}
