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

	leadRepo := repository.NewLeadRepository(db)
	leadSvc := service.NewLeadService(leadRepo)
	leadHandler := handlers.NewLeadHandler(leadSvc)

	propertyRepo := repository.NewPropertyRepository(db)
	propertyService := service.NewPropertyService(propertyRepo)
	propertyHandler := handlers.NewPropertyHandler(propertyService)

	companyRepo := repository.NewCompanyRepository(db)
	companyService := service.NewCompanyService(companyRepo)
	companyHandler := handlers.NewCompanyHandler(companyService)

	agentRepo := repository.NewAgentRepository(db)
	agentService := service.NewAgentService(agentRepo, propertyRepo)
	agentHandler := handlers.NewAgentHandler(agentService)

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)
	messageHandler := handlers.NewMessageHandler(messageService)

	authHandler := handlers.NewAuthHandler(agentService, companyService)

	mux := router.NewRouter(leadHandler, propertyHandler, companyHandler, agentHandler, messageHandler, authHandler)

	// Wrap the router with CORS middleware
	handler := middleware.CorsMiddleware(mux)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
