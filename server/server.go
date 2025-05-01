package server

import (
	"context"
	"fmt"
	"io"
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

	s.mux.HandleFunc("GET /file/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		filename, reader, err := s.store.GetFile(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		defer reader.Close()

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachement;filename=%s", filename))
		io.Copy(w, reader)
	})

	s.mux.HandleFunc("POST /file", func(w http.ResponseWriter, r *http.Request) {
		file, head, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		defer file.Close()

		id, err := s.store.SaveFile(head.Filename, int(head.Size), file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		resp := fmt.Sprintf(`
			<div class="flex flex-col gap-4">
				<p class="w-full text-center">Code for %s</p>
				<p class="font-bold text-green-600 w-full text-center text-5xl">%s</p>
			</div>
		`, head.Filename, id)

		log.Printf("new file uploaded: '%s' of %d bytes, id=%s", head.Filename, head.Size, id)
		w.Write([]byte(resp))
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
