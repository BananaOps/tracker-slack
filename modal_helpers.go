package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

// createProjectDropdown crée un dropdown avec la liste des projets
func createProjectDropdown(blockID, label, initialValue string) (*slack.InputBlock, error) {
	projects, err := GetProjects()
	if err != nil {
		fmt.Printf("Error getting projects for dropdown: %v\n", err)
		// Fallback vers un champ texte
		return createProjectTextInput(blockID, label, initialValue), nil
	}

	if len(projects) == 0 {
		fmt.Println("No projects found, using text input")
		return createProjectTextInput(blockID, label, initialValue), nil
	}

	// Créer les options pour le dropdown
	var options []*slack.OptionBlockObject
	for _, project := range projects {
		option := slack.NewOptionBlockObject(
			project,
			slack.NewTextBlockObject("plain_text", project, false, false),
			nil,
		)
		options = append(options, option)
	}

	// Créer l'élément select
	selectElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject("plain_text", "Sélectionnez un projet", false, false),
		blockID,
		options...,
	)

	// Définir la valeur initiale
	if initialValue != "" {
		// Vérifier que la valeur initiale existe dans la liste
		found := false
		for _, project := range projects {
			if project == initialValue {
				found = true
				break
			}
		}

		if found {
			selectElement.InitialOption = slack.NewOptionBlockObject(
				initialValue,
				slack.NewTextBlockObject("plain_text", initialValue, false, false),
				nil,
			)
		} else {
			// Utiliser le premier projet par défaut
			selectElement.InitialOption = slack.NewOptionBlockObject(
				projects[0],
				slack.NewTextBlockObject("plain_text", projects[0], false, false),
				nil,
			)
		}
	} else if len(projects) > 0 {
		// Pas de valeur initiale, utiliser le premier projet
		selectElement.InitialOption = slack.NewOptionBlockObject(
			projects[0],
			slack.NewTextBlockObject("plain_text", projects[0], false, false),
			nil,
		)
	}

	// Créer l'InputBlock
	inputBlock := slack.NewInputBlock(
		blockID,
		slack.NewTextBlockObject("plain_text", label, false, false),
		nil,
		selectElement,
	)

	fmt.Printf("Created project dropdown with %d options\n", len(projects))
	return inputBlock, nil
}

// createProjectTextInput crée un champ texte pour le projet (fallback)
func createProjectTextInput(blockID, label, initialValue string) *slack.InputBlock {
	textElement := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject("plain_text", "Nom du projet", false, false),
		blockID,
	)

	if initialValue != "" {
		textElement.InitialValue = initialValue
	}

	inputBlock := slack.NewInputBlock(
		blockID,
		slack.NewTextBlockObject("plain_text", label, false, false),
		nil,
		textElement,
	)

	fmt.Println("Created project text input (fallback)")
	return inputBlock
}
