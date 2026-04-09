package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "github.com/tavanovyt/first_project/downloader/internal/adapter/http"
	"github.com/tavanovyt/first_project/downloader/internal/adapter/repository"
	"github.com/tavanovyt/first_project/downloader/internal/config"
	"github.com/tavanovyt/first_project/downloader/internal/usecase"

	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBConn)
	if err != nil {
		log.Fatal("failed to connect to db:", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	repo := repository.NewPostgresRepo(db)
	httpDownloader := httpAdapter.NewHTTPFileDownloader(30 * time.Second) // per-file timeout
	uc := usecase.NewDownloaderUsecase(repo, httpDownloader)
	handler := httpAdapter.NewHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /downloads", handler.CreateDownload)
	mux.HandleFunc("GET /downloads/{id}", handler.GetDownloadStatus)
	mux.HandleFunc("GET /downloads/{id}/files/{file_id}", handler.GetFile)

	var rootHandler http.Handler = mux
	rootHandler = httpAdapter.PanicRecoveryMiddleware(rootHandler)
	rootHandler = httpAdapter.RequestIDMiddleware(rootHandler)
	rootHandler = httpAdapter.LoggingMiddleware(rootHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      rootHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("starting server on %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Println("shutting down gracefully, waiting up to 60 seconds")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
