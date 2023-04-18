package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

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
	server := NewServer(addr)

	log.Logger.Info().Msgf("Starting server on port %s...", port)
	err := server.Start()
	if err != nil {
		log.Logger.Fatal().Msgf("Server died: %v", err)
	}
}

type Server struct {
	httpServer *http.Server
	router     *chi.Mux
}

func NewServer(address string) *Server {
	router := chi.NewRouter()
	s := &Server{
		router: router,
		httpServer: &http.Server{
			Addr:    address,
			Handler: router,
		},
	}
	s.router.Get("/", s.homeHandler)
	s.router.Post("/swipe/{leftorright}/", s.swipeHandler)
	return s
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Hello world endpoint for debugging purposes
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
	bodyLen := len(body)

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
		writeJsonResponse(w, http.StatusCreated, resp, bodyLen)
	case "right":
		writeJsonResponse(w, http.StatusCreated, resp, bodyLen)
	default:
		writeErrorResponse(w, r.Method, http.StatusBadRequest, "not left or right")
	}
}

// Marshal and write a JSON response to the response writer
func writeJsonResponse(w http.ResponseWriter, statusCode int, data interface{}, len int) {
	log.Logger.Debug().Interface("send", data).Send()
	respJson, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "error marshaling JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len))
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
