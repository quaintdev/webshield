package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/quaintdev/webshield/src/internal/dot"
	"github.com/quaintdev/webshield/src/internal/repository"
	"github.com/quaintdev/webshield/src/internal/service"
	"github.com/quaintdev/webshield/src/internal/webserver"
)

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

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
