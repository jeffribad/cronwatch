package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/cronwatch/internal/alert"
	"github.com/user/cronwatch/internal/api"
	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/monitor"
	"github.com/user/cronwatch/internal/scheduler"
)

func main() {
	logger := log.New(os.Stdout, "[cronwatch] ", log.LstdFlags)

	cfgPath := "config.yaml"
	if v := os.Getenv("CRONWATCH_CONFIG"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Fatalf("config: %v", err)
	}

	store, err := history.NewStore(cfg.HistoryPath)
	if err != nil {
		logger.Fatalf("history: %v", err)
	}

	var notifiers []alert.Notifier
	if cfg.WebhookURL != "" {
		notifiers = append(notifiers, alert.NewWebhookNotifier(cfg.WebhookURL))
	}
	notifiers = append(notifiers, &alert.LogNotifier{Logger: logger})
	dispatcher := alert.NewDispatcher(notifiers, logger)

	mon := monitor.New(cfg, store, dispatcher, logger)
	sched := scheduler.New(cfg, mon, logger)

	mux := http.NewServeMux()
	handler := api.NewHandler(mon, store, logger)
	handler.RegisterRoutes(mux)

	addr := ":8080"
	if v := os.Getenv("CRONWATCH_ADDR"); v != "" {
		addr = v
	}
	server := &http.Server{Addr: addr, Handler: mux}

	sched.Start()
	go func() {
		logger.Printf("HTTP server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("shutting down...")

	sched.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("server shutdown: %v", err)
	}
}
