package server

import (
	"context"
	"log"
	"net/http"

	"github.com/jesperkha/loshare/config"
	"github.com/jesperkha/loshare/store"
	"github.com/jesperkha/notifier"
)

type Server struct {
	mux    *http.ServeMux
	config *config.Config
	store  *store.Store
}

func New(config *config.Config, store *store.Store) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		config: config,
		store:  store,
	}

	s.mux.Handle("/", http.FileServer(http.Dir("web")))

	s.mux.HandleFunc("POST /file", func(w http.ResponseWriter, r *http.Request) {
		_, head, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		id, err := s.store.SaveFile(head.Filename, int(head.Size))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		log.Printf("new file uploaded: '%s' of %d bytes", head.Filename, head.Size)
		w.Write([]byte(id))
	})

	return s
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
