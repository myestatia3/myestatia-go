package main

import (
	"log"
	"net/http"

	router "github.com/myestatia/myestatia-go/internal/adapters/input"
	handlers "github.com/myestatia/myestatia-go/internal/adapters/input/handler"
	"github.com/myestatia/myestatia-go/internal/adapters/input/middleware"
	"github.com/myestatia/myestatia-go/internal/application/service"
	entity "github.com/myestatia/myestatia-go/internal/domain/entity"
	database "github.com/myestatia/myestatia-go/internal/infrastructure/database"
	repository "github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"github.com/myestatia/myestatia-go/internal/infrastructure/seed"
	"github.com/myestatia/myestatia-go/internal/infrastructure/storage"
)

func main() {

	cfg := database.LoadConfig()

	if cfg.Host == "" {
		cfg = database.Config{
			Host:     "localhost",
			User:     "user",
			Password: "pass",
			DBName:   "mydb",
			Port:     5432,
			SSLMode:  "disable",
		}
	}

	db := database.InitDB(cfg)

	err := db.AutoMigrate(
		&entity.Lead{},
		&entity.Message{},
		&entity.Summary{},
		&entity.PropertySubtype{},
		&entity.Property{},
		&entity.SystemConfig{},
		&entity.Agent{},
		&entity.Company{},
	)
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	log.Println("Database migrated successfully")

	if err := seed.SeedPropertySubtypes(db); err != nil {
		log.Printf("Error seeding subtypes: %v", err)
	}

	leadRepo := repository.NewLeadRepository(db)
	leadSvc := service.NewLeadService(leadRepo)
	leadHandler := handlers.NewLeadHandler(leadSvc)

	propertyRepo := repository.NewPropertyRepository(db)
	propertyService := service.NewPropertyService(propertyRepo)

	companyRepo := repository.NewCompanyRepository(db)
	companyService := service.NewCompanyService(companyRepo)
	companyHandler := handlers.NewCompanyHandler(companyService)

	agentRepo := repository.NewAgentRepository(db)
	agentService := service.NewAgentService(agentRepo, propertyRepo)
	agentHandler := handlers.NewAgentHandler(agentService)

	// Storage
	storageService := storage.NewLocalStorageService("uploads", "http://localhost:8080/uploads")
	propertyHandler := handlers.NewPropertyHandler(propertyService, agentService, companyService, storageService)

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)
	messageHandler := handlers.NewMessageHandler(messageService)

	authHandler := handlers.NewAuthHandler(agentService, companyService)

	mux := router.NewRouter(leadHandler, propertyHandler, companyHandler, agentHandler, messageHandler, authHandler)

	// Wrap the router with CORS middleware
	// Add static file handler for uploads
	fileServer := http.StripPrefix("/uploads", http.FileServer(http.Dir("uploads")))

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" && len(r.URL.Path) > 8 && r.URL.Path[:8] == "/uploads" {
			fileServer.ServeHTTP(w, r)
			return
		}
		middleware.CorsMiddleware(mux).ServeHTTP(w, r)
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, finalHandler))
}
