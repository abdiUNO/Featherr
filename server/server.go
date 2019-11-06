package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/abdiUNO/featherr/database"

	"github.com/abdiUNO/featherr/config"
	"github.com/abdiUNO/featherr/server/middleware"
	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
	server *http.Server
}

func NewServer() (*Server, error) {
	router := mux.NewRouter()

	database.InitDatabase()

	router.Use(middleware.JwtAuthentication)

	s := &Server{
		router: router,
	}

	s.SetupRoutes()

	return s, nil
}

func (s *Server) ListenAndServe() error {
	cfg := config.GetConfig()

	s.server = &http.Server{
		Addr:    net.JoinHostPort(cfg.AppDomain, cfg.AppPort),
		Handler: handlers.CompressHandler(s.router),
	}

	err := s.server.ListenAndServe()

	fmt.Println("Listening on localhost")

	if err != nil {
		return fmt.Errorf("Could not listen on %s: %v", s.server.Addr, err)
	}

	return nil
}
