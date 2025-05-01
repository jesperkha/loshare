package main

import (
	"log"
	"os"
	"syscall"

	"github.com/jesperkha/loshare/config"
	"github.com/jesperkha/loshare/server"
	"github.com/jesperkha/notifier"
)

func main() {
	notif := notifier.New()

	config := config.Load()
	server := server.New(config)

	go server.ListenAndServe(notif)

	notif.NotifyOnSignal(os.Interrupt, syscall.SIGTERM)
	log.Println("shutdown")
}
