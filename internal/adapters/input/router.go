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
	emailConfigHandler *handler.CompanyEmailConfigHandler,
	googleOAuthHandler *handler.GoogleOAuthHandler,
	passwordResetHandler *handler.PasswordResetHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Password Reset (Public endpoints)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", passwordResetHandler.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/reset-password", passwordResetHandler.ResetPassword)
	mux.HandleFunc("GET /api/v1/auth/validate-reset-token/{token}", passwordResetHandler.ValidateResetToken)

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

	// Company Email Configuration (for MyAccount integration)
	mux.Handle("POST /api/v1/companies/{id}/email-config", protected(emailConfigHandler.CreateEmailConfig))
	mux.Handle("GET /api/v1/companies/{id}/email-config", protected(emailConfigHandler.GetEmailConfig))
	mux.Handle("PUT /api/v1/companies/{id}/email-config", protected(emailConfigHandler.UpdateEmailConfig))
	mux.Handle("DELETE /api/v1/companies/{id}/email-config", protected(emailConfigHandler.DeleteEmailConfig))
	mux.Handle("POST /api/v1/companies/{id}/email-config/test", protected(emailConfigHandler.TestConnection))
	mux.Handle("PATCH /api/v1/companies/{id}/email-config/toggle", protected(emailConfigHandler.ToggleEnabled))

	// Google OAuth2 for Gmail (Public endpoints - no auth required for OAuth flow)
	mux.HandleFunc("GET /api/v1/auth/google/connect", googleOAuthHandler.InitiateOAuth)
	mux.HandleFunc("GET /api/v1/auth/google/callback", googleOAuthHandler.HandleCallback)
	mux.Handle("POST /api/v1/auth/google/disconnect", protected(googleOAuthHandler.DisconnectGmail))

	return mux

}
