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

func New(config *config.Config) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		config: config,
		store:  store.New(),
	}

	s.mux.Handle("/", http.FileServer(http.Dir("web")))

	s.mux.HandleFunc("POST /file", func(w http.ResponseWriter, r *http.Request) {
		_, head, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if head.Size > store.MAX_SIZE {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("rejected '%s': file was too large (%d bytes)", head.Filename, head.Size)
			return
		}

		id := s.store.SaveFile(head.Filename, int(head.Size))
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
