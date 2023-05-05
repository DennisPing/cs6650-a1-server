package main

import (
	"fmt"
	"os"

	"github.com/DennisPing/cs6650-a1-server/log"
	"github.com/DennisPing/cs6650-a1-server/metrics"
	"github.com/DennisPing/cs6650-a1-server/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	datasetName := os.Getenv("AXIOM_DATASET")
	apiToken := os.Getenv("AXIOM_API_TOKEN")
	ingestUrl := "https://api.axiom.co/v1/datasets/%s/ingest"

	if datasetName == "" || apiToken == "" {
		log.Logger.Fatal().Msgf("you forgot to add the Axiom env variables")
	}
	metrics := metrics.NewMetrics(datasetName, apiToken, ingestUrl)

	server := server.NewServer(addr, metrics)

	log.Logger.Info().Msgf("Starting server on port %s...", port)
	err := server.Start()
	if err != nil {
		log.Logger.Fatal().Msgf("server died: %v", err)
	}
}
