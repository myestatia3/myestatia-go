package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/myestatia/myestatia-go/internal/adapters/email/parser"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/email"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

type EmailLeadService struct {
	parserFactory *parser.ParserFactory
	propertyRepo  repository.PropertyRepository
	leadRepo      repository.LeadRepository
	emailConfig   email.Config
}

func NewEmailLeadService(
	propertyRepo repository.PropertyRepository,
	leadRepo repository.LeadRepository,
	emailConfig email.Config,
) *EmailLeadService {
	return &EmailLeadService{
		parserFactory: parser.NewParserFactory(),
		propertyRepo:  propertyRepo,
		leadRepo:      leadRepo,
		emailConfig:   emailConfig,
	}
}

func (s *EmailLeadService) ProcessEmail(ctx context.Context, emailMsg email.ParsedEmail) error {
	subject := emailMsg.Subject
	from := emailMsg.From
	body := emailMsg.Body

	// Step 1: Find appropriate parser
	emailParser, err := s.parserFactory.GetParser(subject, from)
	if err != nil {
		return fmt.Errorf("unsupported email source: %w", err)
	}

	// Step 2: Parse email to extract lead data
	parsedLead, err := emailParser.Parse(subject, body)
	if err != nil {
		return fmt.Errorf("failed to parse email: %w", err)
	}

	log.Printf("[EmailLeadService] Parsed lead from %s: email=%s, ref=%s",
		parsedLead.Source, parsedLead.Email, parsedLead.PropertyReference)

	// Step 3: CRITICAL VALIDATION - Check if property exists
	property, err := s.propertyRepo.FindByReference(ctx, parsedLead.PropertyReference)
	if err != nil {
		return fmt.Errorf("error checking property reference: %w", err)
	}
	if property == nil {
		log.Printf("[EmailLeadService] SKIPPED: Property reference %s not found in database",
			parsedLead.PropertyReference)
		return fmt.Errorf("property reference %s does not exist - email ignored",
			parsedLead.PropertyReference)
	}

	log.Printf("[EmailLeadService] Property %s found (ID: %s)",
		parsedLead.PropertyReference, property.ID)

	// Step 4: Check if lead already exists (duplicate detection)
	existingLead, err := s.leadRepo.FindByEmail(ctx, parsedLead.Email)
	if err != nil {
		return fmt.Errorf("error checking existing lead: %w", err)
	}

	if existingLead != nil {
		// UPDATE existing lead
		return s.updateExistingLead(ctx, existingLead, parsedLead, property)
	}

	// CREATE new lead
	return s.createNewLead(ctx, parsedLead, property)
}

func (s *EmailLeadService) createNewLead(ctx context.Context, parsedLead *entity.ParsedLead, property *entity.Property) error {
	now := time.Now()

	lead := &entity.Lead{
		Name:            parsedLead.Name,
		Email:           parsedLead.Email,
		Phone:           parsedLead.Phone,
		Status:          entity.LeadStatusNew,
		PropertyID:      &property.ID,
		CompanyID:       s.emailConfig.DefaultCompanyID,
		Source:          string(parsedLead.Source),
		Channel:         "email",
		Notes:           fmt.Sprintf("Mensaje inicial: %s", parsedLead.Message),
		LastInteraction: &now,
	}

	if err := s.leadRepo.Create(lead); err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}

	log.Printf("[EmailLeadService] ✓ Created new lead: ID=%s, Email=%s, Property=%s",
		lead.ID, lead.Email, property.Reference)

	return nil
}

// updateExistingLead updates an existing lead with new contact information
func (s *EmailLeadService) updateExistingLead(ctx context.Context, existingLead *entity.Lead, parsedLead *entity.ParsedLead, property *entity.Property) error {
	now := time.Now()

	log.Printf("[EmailLeadService] Lead already exists (ID: %s), updating...", existingLead.ID)

	// Track property history if it changed
	propertyChanged := false
	if existingLead.PropertyID == nil || *existingLead.PropertyID != property.ID {
		propertyChanged = true
		oldPropertyRef := "none"
		if existingLead.PropertyID != nil {
			oldPropertyRef = *existingLead.PropertyID
		}

		historyNote := fmt.Sprintf("\n[%s] Nuevo contacto sobre propiedad %s (anterior: %s). Mensaje: %s",
			now.Format("2006-01-02 15:04"),
			property.Reference,
			oldPropertyRef,
			parsedLead.Message)
		existingLead.Notes += historyNote
		existingLead.PropertyID = &property.ID
	} else {
		// Same property, just update with new message
		messageNote := fmt.Sprintf("\n[%s] Nuevo contacto. Mensaje: %s",
			now.Format("2006-01-02 15:04"),
			parsedLead.Message)
		existingLead.Notes += messageNote
	}

	// Update contact information (in case it changed)
	if parsedLead.Name != "" {
		existingLead.Name = parsedLead.Name
	}
	if parsedLead.Phone != "" {
		existingLead.Phone = parsedLead.Phone
	}

	// Update last interaction timestamp
	existingLead.LastInteraction = &now

	// Update status if it was closed, set it back to contacted
	if existingLead.Status == entity.LeadStatusClosed {
		existingLead.Status = entity.LeadStatusContacted
		log.Printf("[EmailLeadService] Lead status changed from 'closed' to 'contacted'")
	} else if existingLead.Status == entity.LeadStatusNew {
		existingLead.Status = entity.LeadStatusContacted
	}

	// Update source if it changed
	if string(parsedLead.Source) != existingLead.Source {
		existingLead.Source = string(parsedLead.Source)
	}

	if err := s.leadRepo.Update(existingLead); err != nil {
		return fmt.Errorf("failed to update lead: %w", err)
	}

	if propertyChanged {
		log.Printf("[EmailLeadService] ✓ Updated lead with property change: ID=%s, New Property=%s",
			existingLead.ID, property.Reference)
	} else {
		log.Printf("[EmailLeadService] ✓ Updated lead with new contact: ID=%s, Property=%s",
			existingLead.ID, property.Reference)
	}

	return nil
}
