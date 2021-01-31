package web

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/valve"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

// Server is the web server for the Zeta Machine
type Server struct {
	host       string
	port       string
	subdomains []string
	valve      *valve.Valve
	store      *Store
}

// Run reads the configuration from the environment etc., configures routes and
// then listens for requests
func (s *Server) Run() error {
	if err := s.config(); err != nil {
		return err
	}

	r, err := s.routes()
	if err != nil {
		return err
	}

	s.store.Start()

	log.Println("Listening and serving on :" + s.port)
	if http.ListenAndServe(":"+s.port, r) != nil {
	}
	return err
}

func (s *Server) config() error {
	if err := s.checkEnv(); err != nil {
		return err
	}

	s.host = os.Getenv("ZETA_HOSTNAME")
	s.port = os.Getenv("ZETA_PORT")
	s.subdomains = strings.Split(os.Getenv("ZETA_SUBDOMAINS"), ",")
	s.valve = valve.New()

	store, err := NewStore(s.valve)
	if err != nil {
		return err
	}

	s.store = store
	return nil
}

func (s *Server) checkEnv() error {
	godotenv.Load()

	if os.Getenv("ZETA_HOSTNAME") == "" {
		return errors.New("ZETA_HOSTNAME is not set")
	}

	if os.Getenv("ZETA_PORT") == "" {
		return errors.New("ZETA_PORT is not set")
	}

	if os.Getenv("ZETA_SUBDOMAINS") == "" {
		return errors.New("ZETA_SUBDOMAINS is not set")
	}

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not set")
	}

	return nil
}

func (s *Server) routes() (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", s.serveIndex())
	r.Get("/tile/{zoom}/{y}/{x}/", s.serveTile())

	return r, nil
}
