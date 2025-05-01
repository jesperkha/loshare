package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jesperkha/loshare/config"
	"github.com/jesperkha/notifier"
)

type Server struct {
	mux    *http.ServeMux
	config *config.Config
}

func New(config *config.Config) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		config: config,
	}

	fs := http.FileServer(http.Dir("web"))
	s.handle("/", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})

	s.handle("POST /file", func(w http.ResponseWriter, r *http.Request) {
		log.Println("new request!")
		time.Sleep(time.Second * 3)
		w.Write([]byte("hgello"))
	})

	return s
}

func (s *Server) handle(path string, h http.HandlerFunc) {
	s.mux.Handle(path, h)
}

func (s *Server) ListenAndServe(notif *notifier.Notifier) {
	done, finish := notif.Register()

	server := &http.Server{
		Addr:    s.config.Port,
		Handler: s.mux,
	}

	go func() {
		<-done
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
		finish()
	}()

	log.Println("listening on port " + s.config.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Println(err)
	}
}
