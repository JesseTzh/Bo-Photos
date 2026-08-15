package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/besscroft/bophotos/backend/internal/album"
	"github.com/besscroft/bophotos/backend/internal/api"
	"github.com/besscroft/bophotos/backend/internal/asset"
	"github.com/besscroft/bophotos/backend/internal/auth"
	"github.com/besscroft/bophotos/backend/internal/config"
	"github.com/besscroft/bophotos/backend/internal/frontend"
	"github.com/besscroft/bophotos/backend/internal/imageproc"
	"github.com/besscroft/bophotos/backend/internal/jobs"
	"github.com/besscroft/bophotos/backend/internal/media"
	"github.com/besscroft/bophotos/backend/internal/repository"
	"github.com/besscroft/bophotos/backend/internal/site"
	localstorage "github.com/besscroft/bophotos/backend/internal/storage"
	"github.com/besscroft/bophotos/backend/internal/tag"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	runtime, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := repository.Open(ctx, runtime.DatabasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := repository.Migrate(db); err != nil {
		return err
	}

	sessions, stopSessionCleanup := auth.NewSessionManager(db, runtime.CookieSecure)
	defer stopSessionCleanup()
	authService := auth.NewService(auth.NewRepository(db))
	if err := authService.EnsureInitialized(ctx, runtime.InitialPassword); err != nil {
		return err
	}
	authHandler := auth.NewHTTPHandler(
		authService,
		sessions,
	)
	localStorage, err := localstorage.NewLocal(runtime.DataDir)
	if err != nil {
		return err
	}
	assetRepository := asset.NewRepository(db)
	siteRepository := site.NewRepository(db)
	initialSettings, err := siteRepository.Get(ctx)
	if err != nil {
		return err
	}
	processor := imageproc.NewCommandProcessor(nil, imageproc.Commands{}, imageproc.Limits{
		PreviewMaxWidth: initialSettings.PreviewMaxWidth,
		PreviewQuality:  initialSettings.PreviewQuality,
	})
	assetService := asset.NewService(
		assetRepository,
		localStorage,
		processor,
		nil,
		runtime.MaxUploadBytes,
	)
	assetPool := jobs.NewPool(ctx, 2, 64, assetService.Process)
	defer assetPool.Stop()
	assetService.SetQueue(assetPool)
	assetHTTP := asset.NewHTTPHandler(assetService, assetRepository)
	albumRepository := album.NewRepository(db)
	albumHTTP := album.NewHTTPHandler(album.NewService(albumRepository), albumRepository)
	tagRepository := tag.NewRepository(db)
	tagHTTP := tag.NewHTTPHandler(tag.NewService(tagRepository), tagRepository)
	siteHTTP := site.NewHandler(siteRepository, runtime.DataDir)
	cleanup := jobs.NewCleanup(assetRepository, localStorage, time.Hour, 30*24*time.Hour)
	go cleanup.Run(ctx, time.Hour)

	versionedAPI := chi.NewRouter()
	versionedAPI.Mount("/", authHandler.Routes())
	publicAPI := chi.NewRouter()
	publicAPI.Mount("/assets", assetHTTP.PublicRoutes())
	publicAPI.Mount("/albums", albumHTTP.PublicRoutes())
	publicAPI.Mount("/tags", tagHTTP.PublicRoutes())
	siteHTTP.RegisterPublic(publicAPI)
	versionedAPI.Mount("/public", publicAPI)
	adminAPI := chi.NewRouter()
	adminAPI.Mount("/assets", assetHTTP.AdminRoutes(tagHTTP.RegisterAssetRoutes))
	adminAPI.Mount("/albums", albumHTTP.AdminRoutes())
	adminAPI.Mount("/tags", tagHTTP.AdminRoutes())
	siteHTTP.RegisterAdmin(adminAPI)
	versionedAPI.Mount(
		"/admin",
		authHandler.SessionMiddleware(authHandler.ProtectAdministrator(adminAPI)),
	)

	mediaHandler := authHandler.SessionMiddleware(media.NewHandler(
		assetRepository,
		localStorage,
		authHandler.IsAdministrator,
		func(ctx context.Context) bool {
			settings, err := siteRepository.Get(ctx)
			return err == nil && settings.PublicOriginalDownload
		},
	))
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = siteRepository.Cleanup(ctx)
			}
		}
	}()

	server := &http.Server{
		Addr:              runtime.Address,
		Handler:           api.NewRouter(db, frontend.New(runtime.FrontendDir), versionedAPI, mediaHandler),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "address", runtime.Address)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
