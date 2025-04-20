package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dubbie/calculator-api/internal/config"
	"github.com/dubbie/calculator-api/internal/database"
	"github.com/dubbie/calculator-api/internal/domain"
	"github.com/dubbie/calculator-api/internal/handler"
	"github.com/dubbie/calculator-api/internal/service"
	"github.com/dubbie/calculator-api/internal/storage/mysql"
)

func main() {
	fmt.Println("Starting Crafting API server...")

	// 1. Load Configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		return
	}
	fmt.Println("Configuration loaded.")

	// 2. Estabilish Database Connection
	db, err := database.NewDBConnection(cfg)
	if err != nil {
		fmt.Printf("Failed to establish database connection: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database connection established.")

	// 3. Initialize Storage Layer
	itemStore := mysql.NewMySQLItemStore(db)
	craftingMethodStore := mysql.NewMySQLCraftingMethodStore(db)

	// 4. Initialze Service Layer
	itemService := service.NewItemService(itemStore)
	craftingMethodService := service.NewCraftingMethodService(craftingMethodStore)
	// Cast custom list services to the generic ListService interface for items
	itemListService := itemService.(service.ListService[domain.Item, domain.ItemFilters])
	craftingMethodListService := craftingMethodService.(service.ListService[domain.CraftingMethod, domain.CraftingMethodFilters])

	// 5. Setup Router & Handlers
	router := handler.SetupRoutes(cfg, itemService, itemListService, craftingMethodService, craftingMethodListService)
	fmt.Println("Router setup complete.")

	// 6. Create and Configure HTTP Server
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
		// Good practice: Set timeouts to prevent slow-loris attacks
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Start Server in a Goroutine
	go func() {
		fmt.Printf("Server listening on port %s\n", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Error starting server: %v\n", err)
			os.Exit(1)
		}
	}()

	// 8. Graceful Shutdown Handling
	quit := make(chan os.Signal, 1)
	// signal.Notify listens for specified signals (interrupt, terminate)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	<-quit
	fmt.Println("Shutdown signal received, initiating graceful shutdown...")

	// Create a context with a timeout for shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30-second timeout
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server exiting gracefully.")
}
