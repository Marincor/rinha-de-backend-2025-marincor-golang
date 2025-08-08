package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/app"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/app/appinstance"
)

func main() {
	amountOfSignalsToClose := 1
	sigChan := make(chan os.Signal, amountOfSignalsToClose)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	app.ApplicationInit()
	defer func() {
		if err := appinstance.Data.DB.Close(); err != nil {
			log.Print(
				map[string]interface{}{
					"message": "error closing db connection",
					"error":   err,
				},
			)
		} else {
			log.Print("DB connection closed")
		}
	}()

	appinstance.Data.Server = route()

	go app.Setup(appinstance.Data.Config.ServerPort)
	defer func() {
		if err := appinstance.Data.Server.Shutdown(); err != nil {
			log.Print(
				map[string]interface{}{
					"message": "error shutting down server",
					"error":   err,
				},
			)
		} else {
			log.Print("Server closed")
		}
	}()

	<-sigChan
	log.Print("Received signal, shutting down...")

	log.Print("Waiting for tasks to finish...")
}
