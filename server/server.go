package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DennisPing/cs6650-a1-server/log"
	"github.com/DennisPing/cs6650-a1-server/metrics"
	"github.com/DennisPing/cs6650-a1-server/models"
	"github.com/go-chi/chi"
)

type Server struct {
	*chi.Mux
	*http.Server
	*metrics.Metrics
}

func NewServer(address string, metrics *metrics.Metrics) *Server {
	chiRouter := chi.NewRouter()
	s := &Server{
		Mux: chiRouter,
		Server: &http.Server{
			Addr:    address,
			Handler: chiRouter,
		},
		Metrics: metrics,
	}

	s.Get("/health", s.homeHandler)
	s.Post("/swipe/{leftorright}/", s.swipeHandler)
	return s
}

func (s *Server) Start() error {
	ticker := time.NewTicker(5 * time.Second)
	go func() { // Metrics goroutine
		for range ticker.C {
			err := s.SendMetrics()
			if err != nil {
				log.Logger.Error().Msgf("unable to send metrics to Axiom: %v", err)
			}
		}
	}()
	return s.ListenAndServe()
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
		s.IncrementThroughput()
		writeJsonResponse(w, http.StatusCreated, resp)
	case "right":
		s.IncrementThroughput()
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
