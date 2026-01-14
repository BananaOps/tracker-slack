package main

import (
	"log/slog"

	"github.com/slack-go/slack"
)

// createProjectDropdown crée un dropdown avec la liste des projets
func createProjectDropdown(blockID, label, initialValue string) (*slack.InputBlock, error) {
	projects, err := GetProjects()
	if err != nil {
		logger.Error("Failed to get projects for dropdown", slog.Any("error", err))
		// Fallback vers un champ texte
		return createProjectTextInput(blockID, label, initialValue), nil
	}

	if len(projects) == 0 {
		logger.Warn("No projects found, using text input")
		return createProjectTextInput(blockID, label, initialValue), nil
	}

	// Si plus de 100 projets, utiliser un external select avec recherche
	if len(projects) > 100 {
		logger.Debug("Using external select for projects", slog.Int("count", len(projects)))
		return createProjectExternalSelect(blockID, label, initialValue), nil
	}

	// Créer les options pour le dropdown statique
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
		"project", // Action ID
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
		}
	}

	// Créer l'InputBlock
	inputBlock := slack.NewInputBlock(
		blockID,
		slack.NewTextBlockObject("plain_text", label, false, false),
		nil,
		selectElement,
	)

	logger.Debug("Created static project dropdown", slog.Int("options", len(projects)))
	return inputBlock, nil
}

// createProjectExternalSelect crée un select avec recherche externe pour plus de 100 projets
func createProjectExternalSelect(blockID, label, initialValue string) *slack.InputBlock {
	// Créer un external select qui permettra la recherche
	selectElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeExternal,
		slack.NewTextBlockObject("plain_text", "Rechercher un projet...", false, false),
		"project", // Action ID - doit correspondre dans handleOptionLoadEndpoint
	)

	// Définir la valeur initiale si elle existe
	if initialValue != "" {
		selectElement.InitialOption = slack.NewOptionBlockObject(
			initialValue,
			slack.NewTextBlockObject("plain_text", initialValue, false, false),
			nil,
		)
	}

	// Nombre minimum de caractères avant de déclencher la recherche
	minQueryLength := 2
	selectElement.MinQueryLength = &minQueryLength

	inputBlock := slack.NewInputBlock(
		blockID,
		slack.NewTextBlockObject("plain_text", label, false, false),
		nil,
		selectElement,
	)

	logger.Debug("Created external select for project search")
	return inputBlock
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

	logger.Debug("Created project text input (fallback)")
	return inputBlock
}
