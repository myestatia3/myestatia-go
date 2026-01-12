package router

import (
	"net/http"

	"github.com/myestatia/myestatia-go/internal/adapters/input/handler"
	"github.com/myestatia/myestatia-go/internal/adapters/input/middleware"
)

func NewRouter(
	leadHandler *handler.LeadHandler,
	propertyHandler *handler.PropertyHandler,
	companyHandler *handler.CompanyHandler,
	agentHandler *handler.AgentHandler,
	messageHandler *handler.MessageHandler,
	authHandler *handler.AuthHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Helper to protect routes
	protected := func(h http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(h)
	}

	// CRUD Leads
	mux.Handle("POST /api/v1/leads", protected(leadHandler.CreateLead))
	mux.Handle("GET /api/v1/leads", protected(leadHandler.GetAllLeads))
	mux.Handle("GET /api/v1/leads/{id}", protected(leadHandler.GetLeadByID))
	mux.Handle("PUT /api/v1/leads/{id}", protected(leadHandler.UpdateLead))
	mux.Handle("DELETE /api/v1/leads/{id}", protected(leadHandler.DeleteLead))
	mux.Handle("GET /api/v1/leads/bycompany/{companyId}", protected(leadHandler.GetLeadByCompanyId))
	mux.Handle("GET /api/v1/leads/byproperty/{propertyId}", protected(leadHandler.GetLeadByPropertyId))

	//Property search filters (Public? Or Protected? Let's protect for now to enforce users)
	mux.Handle("GET /api/v1/properties/search", protected(propertyHandler.SearchProperties))

	// Public Property access
	mux.HandleFunc("GET /api/v1/public/properties/", propertyHandler.GetPublicPropertyByID)

	// CRUD Property
	mux.Handle("POST /api/v1/properties", protected(propertyHandler.CreateProperty))
	mux.Handle("GET /api/v1/properties", protected(propertyHandler.GetAllProperties))
	mux.Handle("GET /api/v1/properties/{id}", protected(propertyHandler.GetPropertyByID))
	mux.Handle("PUT /api/v1/properties/{id}", protected(propertyHandler.UpdateProperty))
	mux.Handle("DELETE /api/v1/properties/{id}", protected(propertyHandler.DeleteProperty))
	mux.Handle("GET /api/v1/properties/company/{company_id}", protected(propertyHandler.GetPropertiesByCompany))
	mux.Handle("GET /api/v1/property-subtypes", protected(propertyHandler.ListSubtypes))

	//CRUD Company
	mux.Handle("POST /api/v1/companies", protected(companyHandler.CreateCompany))
	mux.Handle("GET /api/v1/companies", protected(companyHandler.GetAllCompanies))
	mux.Handle("GET /api/v1/companies/{id}", protected(companyHandler.GetCompanyByID))
	mux.Handle("PUT /api/v1/companies/{id}", protected(companyHandler.UpdateCompany))
	mux.Handle("DELETE /api/v1/companies/{id}", protected(companyHandler.DeleteCompany))

	//CRUD Agent
	mux.Handle("POST /api/v1/agents", protected(agentHandler.CreateAgent))
	mux.Handle("GET /api/v1/agents", protected(agentHandler.GetAllAgents))
	mux.Handle("GET /api/v1/agents/{id}", protected(agentHandler.GetAgentByID))
	mux.Handle("PUT /api/v1/agents/{id}", protected(agentHandler.UpdateAgent))
	mux.Handle("DELETE /api/v1/agents/{id}", protected(agentHandler.DeleteAgent))

	// Conversations
	mux.Handle("GET /api/v1/lead/{id}/conversations", protected(messageHandler.GetConversations))
	mux.Handle("POST /api/v1/conversations/{leadId}/messages", protected(messageHandler.SendMessage))

	return mux
}
