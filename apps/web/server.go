package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"zetamachine/pkg/zeta"

	"github.com/foolin/goview"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type server struct {
	host       string
	port       string
	luts       []*zeta.LUT
	subdomains []string
}

func (s *server) config() error {
	s.subdomains = strings.Split(os.Getenv("ZETA_SUBDOMAINS"), ",")
	return nil
}

func (s *server) run() error {
	s.config()
	r, _ := s.routes()

	return http.ListenAndServe(":"+s.port, r)
}

func (s *server) loadLuts() error {
	luts := make([]*zeta.LUT, 6)
	luts[0] = &zeta.LUT{ID: 0, FName: "CL100000.bmp", Min: .975 - .025i, Max: 1.025 + .025i, Res: 100000}
	luts[1] = &zeta.LUT{ID: 1, FName: "CL010000.bmp", Min: .75 - .25i, Max: 1.25 + .25i, Res: 10000}
	luts[2] = &zeta.LUT{ID: 2, FName: "CL001000.bmp", Min: -1.5 - 2.5i, Max: 3.5 + 2.5i, Res: 1000}
	luts[3] = &zeta.LUT{ID: 3, FName: "CL000100.bmp", Min: -24 - 25i, Max: 26 + 25i, Res: 100}
	luts[4] = &zeta.LUT{ID: 4, FName: "CL000010.bmp", Min: -249 - 250i, Max: 251 + 250i, Res: 10}
	luts[5] = &zeta.LUT{ID: 5, FName: "CL000001.bmp", Min: -2499 - 2500i, Max: 2501 + 2500i, Res: 1}

	s.luts = luts

	fmt.Print("Loading LUTs...")
	start := time.Now()

	wg := &sync.WaitGroup{}
	wg.Add(len(luts))
	for i := range luts {
		go func(id int) {
			l := s.luts[id]
			if err := l.Load(); err != nil {
				log.Fatal("error loading luts", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	sort.Slice(luts, func(i, j int) bool { return luts[i].ID < luts[j].ID })
	fmt.Println("in", time.Since(start).Milliseconds(), "ms")

	return nil
}

func (s *server) routes() (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		goview.DefaultConfig.DisableCache = true
		err := goview.Render(w, http.StatusOK, "index.html", goview.M{
			"host":       s.host + ":" + s.port,
			"subdomains": strings.Join(s.subdomains, ""),
		})

		if err != nil {
			fmt.Fprintf(w, "Render index error: %v!", err)
		}
	})

	r.Get("/tile/{zoom}/{y}/{x}/", handleTileReq(s))
	r.Get("/generate/{zoom}/{y}/{x}/", handleGenerate(s.luts))

	return r, nil
}
