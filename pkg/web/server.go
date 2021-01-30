package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"zetamachine/pkg/zeta"

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
	luts       []*zeta.LUT
	valve      *valve.Valve
}

// Run reads the configuration from the environment etc., configures routes and
// then listens for requests
func (s *Server) Run() error {
	if err := s.config(); err != nil {
		return err
	}

	if os.Getenv("ZETA_USE_LUTS") != "" {
		start := time.Now()

		fmt.Print("Loading LUTs ... ")
		luts, err := zeta.LoadLUTs()
		if err != nil {
			log.Println("Continuing without lookup tables: \n\t", err.Error())
		}
		s.luts = luts
		fmt.Println("in", time.Since(start).Milliseconds(), "ms")
	}

	r, err := s.routes()
	if err != nil {
		return err
	}

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

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 8

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

	if os.Getenv("ZETA_TILE_GENERATOR_URL") == "" {
		return errors.New("ZETA_TILE_GENERATOR_URL is not set")
	}

	if os.Getenv("ZETA_SUBDOMAINS") == "" {
		return errors.New("ZETA_SUBDOMAINS is not set")
	}

	if os.Getenv("ZETA_NSQLOOKUP") == "" {
		return errors.New("ZETA_NSQLOOKUP is not set")
	}

	if os.Getenv("ZETA_GENERATE_VIA") == "" {
		return errors.New("ZETA_GENERATE_VIA is not set")
	}

	if os.Getenv("ZETA_DEFAULT_ZOOM") == "" {
		return errors.New("ZETA_DEFAULT_ZOOM is not set")
	}

	if os.Getenv("ZETA_DEFAULT_REAL") == "" {
		return errors.New("ZETA_DEFAULT_REAL is not set")
	}

	if os.Getenv("ZETA_DEFAULT_IMAG") == "" {
		return errors.New("ZETA_DEFAULT_IMAG is not set")
	}

	return nil
}

func (s *Server) routes() (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", s.serveIndex())
	r.Get("/tile/{zoom}/{y}/{x}/", s.serveTile())
	r.Post("/generate/", s.generateTile())

	return r, nil
}
