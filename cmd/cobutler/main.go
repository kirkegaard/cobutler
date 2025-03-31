package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/kirkegaard/cobutler/pkg/cobutler/api"
	"github.com/kirkegaard/cobutler/pkg/cobutler/models"
)

func main() {
	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize database connection with high performance settings
	dbFile := "brain.db"
	logger.Info("Initializing brain", "database", dbFile)

	// Create the brain - this will automatically use optimized settings
	brain, err := models.NewBrain(dbFile)
	if err != nil {
		logger.Error("Failed to initialize brain", "error", err)
		os.Exit(1)
	}
	defer brain.Close()

	// Set up API handler with ultra-fast response method
	handler := api.NewHandler(brain)

	// Configure and start HTTP server
	http.HandleFunc("/predict", handler.Predict)
	http.HandleFunc("/learn", handler.Learn)

	logger.Info("Starting server", "address", ":8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Server failed", "error", err)
	}
}
