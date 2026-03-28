package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/application/command"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/application/query"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/infrastructure/postgres"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/infrastructure/pubsub"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/interfaces/http/handler"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, getDatabaseURL())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewPostgresRepository(pool)
	templateRepo := postgres.NewPostgresTemplateRepository(pool)
	gwpRepo := postgres.NewGWPRepository(pool)

	var publisher *pubsub.Publisher
	topicURL := os.Getenv("PUBSUB_TOPIC_URL")
	if topicURL != "" {
		topic, err := pubsub.OpenTopic(ctx, topicURL)
		if err != nil {
			log.Printf("Warning: failed to open pubsub topic: %v", err)
		} else {
			defer topic.Shutdown(ctx)
			publisher = pubsub.NewPublisher(topic)
		}
	} else {
		log.Println("Warning: PUBSUB_TOPIC_URL not set, events will not be published")
	}

	cmdHandler := command.NewInventoryCommandHandler(repo, templateRepo, publisher)
	queryHandler := query.NewInventoryQueryHandler(repo, templateRepo, gwpRepo)

	invHandler := handler.NewInventoryHandler(cmdHandler, queryHandler)
	tplHandler := handler.NewTemplateHandler(cmdHandler, queryHandler)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", handler.HandleHealthz)
	mux.HandleFunc("/readyz", handler.HandleReadyz)

	mux.HandleFunc("/v1/inventories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			invHandler.CreateInventory(w, r)
			return
		}
		if r.Method == http.MethodGet {
			invHandler.ListInventories(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v1/inventories/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/v1/inventories/dashboard" {
			invHandler.GetDashboard(w, r)
			return
		}

		if strings.HasSuffix(path, "/summary") {
			invHandler.GetSummary(w, r)
			return
		}

		if strings.Contains(path, "/state") {
			invHandler.TransitionState(w, r)
			return
		}

		if strings.Contains(path, "/evidences") {
			invHandler.AddEvidence(w, r)
			return
		}

		if strings.Contains(path, "/reliability-job") {
			invHandler.StoreReliabilityJob(w, r)
			return
		}

		if strings.Contains(path, "/variables") {
			invHandler.FillVariables(w, r)
			return
		}

		if r.Method == http.MethodGet {
			invHandler.GetInventory(w, r)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v1/emission-templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			tplHandler.CreateTemplate(w, r)
			return
		}
		if r.Method == http.MethodGet {
			tplHandler.ListTemplates(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	port := getPort()
	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getDatabaseURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://lowcarbon:lowcarbon@localhost:5432/lowcarbon_inventory?sslmode=disable"
	}
	return dbURL
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	return fmt.Sprintf(":%s", port)
}