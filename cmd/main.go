package main

import (
	"log"
	"net/http"

	router "bitbucket.org/statia/server/internal/adapters/input"
	handlers "bitbucket.org/statia/server/internal/adapters/input/handler"
	"bitbucket.org/statia/server/internal/application/service"
	entity "bitbucket.org/statia/server/internal/domain/entity"
	database "bitbucket.org/statia/server/internal/infrastructure/database"
	repository "bitbucket.org/statia/server/internal/infrastructure/repository"
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

	mux := router.NewRouter(leadHandler, propertyHandler, companyHandler, agentHandler)

	addr := ":3000"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
