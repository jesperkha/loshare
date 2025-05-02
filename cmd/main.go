package main

import (
	"log"
	"os"
	"syscall"

	"github.com/jesperkha/loshare/config"
	"github.com/jesperkha/loshare/server"
	"github.com/jesperkha/loshare/store"
	"github.com/jesperkha/notifier"
)

func main() {
	notif := notifier.New()
	config := config.Load()

	store := store.New(config)
	store.Init()

	server := server.New(config, store)

	go server.ListenAndServe(notif)
	go store.Run(notif)

	notif.NotifyOnSignal(os.Interrupt, syscall.SIGTERM)
	log.Println("shutdown")
}
