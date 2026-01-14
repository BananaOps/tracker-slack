package main

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

// InitLogger initialise le logger global avec un format JSON
func InitLogger() {
	// Déterminer le niveau de log depuis l'environnement (défaut: INFO)
	logLevel := os.Getenv("LOG_LEVEL")
	var level slog.Level

	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN", "WARNING":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Créer un handler JSON avec le niveau configuré
	opts := &slog.HandlerOptions{
		Level: level,
		// Ajouter la source (fichier:ligne) pour faciliter le debug
		AddSource: level == slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)

	// Définir comme logger par défaut
	slog.SetDefault(logger)

	logger.Info("Logger initialized",
		slog.String("level", level.String()),
		slog.Bool("add_source", opts.AddSource),
	)
}

// GetLogger retourne l'instance du logger global
func GetLogger() *slog.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}
