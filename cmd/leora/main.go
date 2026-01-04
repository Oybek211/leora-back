package main

import (
    "log"

    "github.com/leora/leora-server/internal/app"
)

func main() {
    application, err := app.NewApplication()
    if err != nil {
        log.Fatalf("failed to initialize application: %v", err)
    }

    if err := application.Start(); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
