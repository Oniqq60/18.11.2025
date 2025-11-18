package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalrouter "github.com/example/2025-11-18/internal"
	"github.com/example/2025-11-18/internal/handlers"
	"github.com/example/2025-11-18/internal/service/checker"
	"github.com/example/2025-11-18/internal/service/report"
	"github.com/example/2025-11-18/internal/storage"
)

func main() {
	dataDir := "./data"
	store, err := storage.NewFileStorage(dataDir)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	checkerSvc := checker.New(store, 10*time.Second)
	reportGen := report.NewGenerator()

	submitHandler := handlers.NewSubmitHandler(store, checkerSvc)
	reportHandler := handlers.NewReportHandler(store, reportGen)

	handler := loggingMiddleware(internalrouter.NewRouter(submitHandler, reportHandler))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("сервер запущен на %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	waitForShutdown(server)
}

func waitForShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		log.Print("сервер остановлен корректно")
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
