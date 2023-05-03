package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/DennisPing/cs6650-a1-server/log"
	"github.com/DennisPing/cs6650-a1-server/models"
	"github.com/go-chi/chi"
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
	metrics := Metrics{
		datasetName: datasetName,
		apiToken:    apiToken,
		ingestUrl:   ingestUrl,
	}

	server := NewServer(addr, metrics)

	log.Logger.Info().Msgf("Starting server on port %s...", port)
	err := server.Start()
	if err != nil {
		log.Logger.Fatal().Msgf("server died: %v", err)
	}
}

type Server struct {
	*chi.Mux
	httpServer *http.Server
	throughput uint64
	mutex      sync.Mutex
	metrics    Metrics
}

type Metrics struct {
	datasetName string
	apiToken    string
	ingestUrl   string
}

func NewServer(address string, metrics Metrics) *Server {
	chiRouter := chi.NewRouter()
	s := &Server{
		Mux: chiRouter,
		httpServer: &http.Server{
			Addr:    address,
			Handler: chiRouter,
		},
		metrics: metrics,
	}

	s.Get("/health", s.homeHandler)
	s.Post("/swipe/{leftorright}/", s.swipeHandler)
	return s
}

func (s *Server) Start() error {
	ticker := time.NewTicker(5 * time.Second)
	go func() { // Metrics goroutine
		for range ticker.C {
			err := s.sendMetrics()
			if err != nil {
				log.Logger.Error().Msgf("unable to send metrics to Axiom: %v", err)
			}
		}
	}()
	return s.httpServer.ListenAndServe()
}

// Health endpoint
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world!"))
}

// Handle swipe left or right
func (s *Server) swipeHandler(w http.ResponseWriter, r *http.Request) {
	leftorright := chi.URLParam(r, "leftorright")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "bad request")
		return
	}

	var reqBody models.SwipeRequest
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "bad request")
		return
	}

	resp := models.SwipeResponse{
		Message: fmt.Sprintf("you swiped %s", leftorright),
	}
	// left and right do the same thing for now
	switch leftorright {
	case "left":
		s.incrementThroughput()
		writeJsonResponse(w, http.StatusCreated, resp)
	case "right":
		s.incrementThroughput()
		writeJsonResponse(w, http.StatusCreated, resp)
	default:
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "not left or right")
	}
}

// Marshal and write a JSON response to the response writer
func writeJsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	log.Logger.Debug().Interface("send", data).Send()
	respJson, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "error marshaling JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(respJson)))
	w.WriteHeader(statusCode)
	w.Write(respJson)
}

// Write an HTTP error to the response writer
func writeErrorResponse(w http.ResponseWriter, method string, statusCode int, message string) {
	log.Logger.Error().
		Str("method", method).
		Int("code", statusCode).
		Msg(message)
	http.Error(w, message, statusCode)
}

// Gathering metrics ********************************************************************

func (s *Server) incrementThroughput() {
	s.mutex.Lock()
	s.throughput++
	s.mutex.Unlock()
}

func (s *Server) getThroughput() uint64 {
	s.mutex.Lock()
	throughput := s.throughput
	s.throughput = 0
	s.mutex.Unlock()
	return throughput
}

func (s *Server) sendMetrics() error {
	throughput := s.getThroughput()
	payload := models.AxiomPayload{
		Time:       time.Now().Format(time.RFC3339Nano),
		Throughput: throughput,
	}

	jsonPayload, err := json.Marshal([]models.AxiomPayload{payload})
	if err != nil {
		return err
	}

	url := fmt.Sprintf(s.metrics.ingestUrl, s.metrics.datasetName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.metrics.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%s", resp.Status)
	}

	return nil
}
