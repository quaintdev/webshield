package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/quaintdev/webshield/src/internal/dot"
	"github.com/quaintdev/webshield/src/internal/repository"
	"github.com/quaintdev/webshield/src/internal/service"
	"github.com/quaintdev/webshield/src/internal/webserver"
)

func initLogger() {
	var level slog.Level

	switch os.Getenv("LOGGING") {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// Replace with just filename:line
					return slog.String("source",
						fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line))
				}
			}
			return a
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	initLogger()

	//init repositories

	domainDataStore := repository.NewDomainDataSTore()
	domainDataRepo := repository.DomainDataRepository(domainDataStore)

	configService := service.NewApplicationConfigService("config.json")
	configService.LoadDomainData(domainDataRepo)

	dataStore, err := repository.NewBoltDataStore("user-data.db")
	if err != nil {
		slog.Error("error while opening database", "error", err)
		return
	}
	defer dataStore.Close()

	settingsRepo := repository.SettingsRepository(dataStore)

	//init services
	serverSelector := service.NewDNSServerSelector(configService.GetDNSServers())
	filteringService := service.NewFilteringService(settingsRepo, domainDataRepo)
	dnsService := service.NewDNSService(serverSelector, filteringService, configService)
	userService := service.NewDataMgmtService(settingsRepo, configService)

	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	server := webserver.NewWebServer(userService, dnsService)
	wg.Add(1)
	go server.Start(&wg)

	if os.Getenv("DOT_SERVER_DISABLED") != "true" {
		dotServer := dot.NewDotServer(dnsService)
		wg.Add(1)
		go dotServer.Start(ctx, &wg, configService)
	}

	select {
	case <-ctx.Done():
		slog.Info("Context canceled, shutting down DNS server")
	case sig := <-signalCh:
		slog.Info("Received signal, shutting down DNS server", "signal", sig)
		cancel()
		server.Shutdown(ctx)
	}

	wg.Wait()
}
