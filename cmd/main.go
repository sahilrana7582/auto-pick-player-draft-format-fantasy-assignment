package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"auto-pick-player-draft-format-fantasy-assignment/internal/db"
	"auto-pick-player-draft-format-fantasy-assignment/internal/handler"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
	"auto-pick-player-draft-format-fantasy-assignment/internal/service"
	"auto-pick-player-draft-format-fantasy-assignment/internal/workers"
)

// How often the auto-pick worker checks for timed-out picks.
const autoPickInterval = 2 * time.Second

func main() {
	// ctx is cancelled on Ctrl+C / SIGTERM -> stops the worker and the server.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load .env into the process env. Missing file is fine (e.g. env set by Docker/CI).
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set (see .env.example)")
	}
	port := getEnv("PORT", "8080")

	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	// Wire layers: repo -> service -> handler.
	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	leagueRepo := repository.NewLeagueRepository(pool)
	leagueSvc := service.NewLeagueService(leagueRepo)
	leagueHandler := handler.NewLeagueHandler(leagueSvc)

	entryRepo := repository.NewEntryRepository(pool)
	entrySvc := service.NewEntryService(entryRepo, userRepo, leagueRepo)
	entryHandler := handler.NewEntryHandler(entrySvc)

	rosterRepo := repository.NewRosterRepository(pool)
	rosterSvc := service.NewRosterService(rosterRepo, entryRepo, leagueRepo)
	rosterHandler := handler.NewRosterHandler(rosterSvc)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)
	leagueHandler.RegisterRoutes(mux)
	entryHandler.RegisterRoutes(mux)
	rosterHandler.RegisterRoutes(mux)

	// Auto-pick worker runs in the background alongside the API.
	autoPickWorker := workers.New(pool, &sync.Map{})
	go autoPickWorker.Run(ctx, autoPickInterval)
	log.Printf("auto-pick worker started (every %s)", autoPickInterval)

	server := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		log.Printf("listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Block until Ctrl+C, then shut down: ctx is already cancelled, so the worker
	// stops on its next loop; give in-flight HTTP requests up to 5s to finish.
	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}

// getEnv returns the env var or a fallback if it's not set.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
