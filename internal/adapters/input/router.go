package router

import (
	"net/http"

	"bitbucket.org/statia/server/internal/adapters/input/handler"
)

func NewRouter(leadHandler *handler.LeadHandler, propertyHandler *handler.PropertyHandler, companyHandler *handler.CompanyHandler, agentHandler *handler.AgentHandler) http.Handler {
	mux := http.NewServeMux()

	// CRUD Leads
	mux.HandleFunc("POST /api/v1/leads", leadHandler.CreateLead)
	mux.HandleFunc("GET /api/v1/leads", leadHandler.GetAllLeads)
	mux.HandleFunc("GET /api/v1/leads/{id}", leadHandler.GetLeadByID)
	mux.HandleFunc("PUT /api/v1/leads/{id}", leadHandler.UpdateLead)
	mux.HandleFunc("DELETE /api/v1/leads/{id}", leadHandler.DeleteLead)
	mux.HandleFunc("GET /api/v1/leads/bycompany/{companyId}", leadHandler.GetLeadByCompanyId)
	mux.HandleFunc("GET /api/v1/leads/byproperty/{propertyId}", leadHandler.GetLeadByPropertyId)

	//Property search filters
	mux.HandleFunc("GET /api/v1/properties/search", propertyHandler.SearchProperties)
	// CRUD Property
	mux.HandleFunc("POST /api/v1/properties", propertyHandler.CreateProperty)
	mux.HandleFunc("GET /api/v1/properties", propertyHandler.GetAllProperties)
	mux.HandleFunc("GET /api/v1/properties/{id}", propertyHandler.GetPropertyByID)
	mux.HandleFunc("PUT /api/v1/properties/{id}", propertyHandler.UpdateProperty)
	mux.HandleFunc("DELETE /api/v1/properties/{id}", propertyHandler.DeleteProperty)
	mux.HandleFunc("GET /api/v1/properties/company/{company_id}", propertyHandler.GetPropertiesByCompany)

	//CRUD Company

	mux.HandleFunc("POST /api/v1/companies", companyHandler.CreateCompany)
	mux.HandleFunc("GET /api/v1/companies", companyHandler.GetAllCompanies)
	mux.HandleFunc("GET /api/v1/companies/{id}", companyHandler.GetCompanyByID)
	mux.HandleFunc("PUT /api/v1/companies/{id}", companyHandler.UpdateCompany)
	mux.HandleFunc("DELETE /api/v1/companies/{id}", companyHandler.DeleteCompany)

	//CRUD Agent

	mux.HandleFunc("POST /api/v1/agents", agentHandler.CreateAgent)
	mux.HandleFunc("GET /api/v1/agents", agentHandler.GetAllAgents)
	mux.HandleFunc("GET /api/v1/agents/{id}", agentHandler.GetAgentByID)
	mux.HandleFunc("PUT /api/v1/agents/{id}", agentHandler.UpdateAgent)
	mux.HandleFunc("DELETE /api/v1/agents/{id}", agentHandler.DeleteAgent)

	return mux
}
