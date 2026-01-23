package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

type PresentationService struct {
	leadRepo     repository.LeadRepository
	propertyRepo repository.PropertyRepository
	agentRepo    repository.AgentRepository
	companyRepo  repository.CompanyRepository
	jwtSecret    string
}

func NewPresentationService(
	leadRepo repository.LeadRepository,
	propertyRepo repository.PropertyRepository,
	agentRepo repository.AgentRepository,
	companyRepo repository.CompanyRepository,
	jwtSecret string,
) *PresentationService {
	return &PresentationService{
		leadRepo:     leadRepo,
		propertyRepo: propertyRepo,
		agentRepo:    agentRepo,
		companyRepo:  companyRepo,
		jwtSecret:    jwtSecret,
	}
}

// GenerateToken for a presentation
func (s *PresentationService) GenerateToken(leadID string, propertyIDs []string) (string, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()

	claims := jwt.MapClaims{
		"leadId":      leadID,
		"propertyIds": propertyIDs,
		"exp":         expiresAt,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *PresentationService) ValidateToken(tokenString string) (*entity.PresentationToken, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || token == nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	leadID, ok := claims["leadId"].(string)
	if !ok {
		return nil, errors.New("invalid leadId in token")
	}

	propertyIDsInterface, ok := claims["propertyIds"].([]any)
	if !ok {
		return nil, errors.New("invalid propertyIds in token")
	}

	propertyIDs := make([]string, len(propertyIDsInterface))
	for i, v := range propertyIDsInterface {
		propertyIDs[i], ok = v.(string)
		if !ok {
			return nil, errors.New("invalid property ID in token")
		}
	}

	expiresAt, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid exp in token")
	}

	return &entity.PresentationToken{
		LeadID:      leadID,
		PropertyIDs: propertyIDs,
		ExpiresAt:   int64(expiresAt),
	}, nil
}

func (s *PresentationService) GetPresentation(ctx context.Context, tokenString string) (*entity.Presentation, error) {
	tokenData, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if time.Now().Unix() > tokenData.ExpiresAt {
		return nil, errors.New("presentation has expired")
	}

	lead, err := s.leadRepo.FindByID(tokenData.LeadID)
	if err != nil {
		return nil, fmt.Errorf("lead not found: %w", err)
	}
	properties := make([]entity.Property, 0, len(tokenData.PropertyIDs))
	for _, propID := range tokenData.PropertyIDs {
		prop, err := s.propertyRepo.FindByID(propID)
		if err != nil {
			continue
		}
		if prop != nil {
			properties = append(properties, *prop)
		}
	}

	contactPhone := ""
	if lead.AssignedAgentID != nil && *lead.AssignedAgentID != "" {
		agent, err := s.agentRepo.FindByID(*lead.AssignedAgentID)
		if err == nil && agent != nil && agent.Phone != "" {
			contactPhone = agent.Phone
		}
	}

	if contactPhone == "" && lead.CompanyID != "" {
		company, err := s.companyRepo.FindByID(lead.CompanyID)
		if err == nil && company != nil {
			contactPhone = company.Phone1
		}
	}

	return &entity.Presentation{
		Lead:         lead,
		Properties:   properties,
		ContactPhone: contactPhone,
	}, nil
}

func (s *PresentationService) GetMatchingProperties(ctx context.Context, leadID string) ([]entity.PropertyMatch, error) {
	lead, err := s.leadRepo.FindByID(leadID)
	if err != nil {
		return nil, fmt.Errorf("lead not found: %w", err)
	}

	slog.Debug("GetMatchingProperties - Lead ID: %s, CompanyID: %s", leadID, lead.CompanyID)

	allProperties, err := s.propertyRepo.FindAllByCompanyID(ctx, lead.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch properties: %w", err)
	}

	slog.Debug("GetMatchingProperties found properties", "count", len(allProperties), "company_id", lead.CompanyID)

	matches := make([]entity.PropertyMatch, 0, len(allProperties))

	for _, prop := range allProperties {
		isInquired := lead.PropertyID != nil && *lead.PropertyID == prop.ID
		isDismissed := isInquired && (lead.Status == entity.LeadStatusClosed ||
			lead.Status == entity.LeadStatusDismissed ||
			lead.Status == entity.LeadStatusRejected)

		// Cuando tengamos el algoritmo del %match esto será real
		matchPercent := s.calculateMatchPercentage(lead, &prop)

		matches = append(matches, entity.PropertyMatch{
			Property:     &prop,
			MatchPercent: matchPercent,
			IsInquired:   isInquired,
			IsDismissed:  isDismissed,
		})
	}

	slog.Debug("GetMatchingProperties - Created matches", "count", len(matches))

	// Sort: Primero las interesadas por el lead, luego las del matching y que no estén descartadas
	sortedMatches := s.sortPropertyMatches(matches)

	slog.Debug("GetMatchingProperties - Returning sorted matches", "count", len(sortedMatches))

	return sortedMatches, nil
}

// calculateMatchPercentage is a placeholder algorithm for calculating match percentage
func (s *PresentationService) calculateMatchPercentage(lead *entity.Lead, property *entity.Property) int {
	// Placeholder: return a random-ish percentage based on property characteristics
	// TODO: Implement actual matching algorithm later

	matchScore := 50 // Base score

	// Budget matching
	if lead.Budget > 0 && property.Price > 0 {
		priceDiff := abs(property.Price - lead.Budget)
		budgetMatchPercent := max(0, 100-int(priceDiff/lead.Budget*100))
		matchScore += budgetMatchPercent / 4
	}

	// Zone matching
	if lead.Zone != "" && property.Zone != "" && lead.Zone == property.Zone {
		matchScore += 20
	}

	// Property type matching (basic string comparison)
	if lead.PropertyType != "" && string(property.Type) != "" {
		// Simple contains check
		matchScore += 10
	}

	// Cap at 100
	if matchScore > 100 {
		matchScore = 100
	}

	return matchScore
}

// sortPropertyMatches sorts matches by criteria: inquired + not dismissed first, then by match %
func (s *PresentationService) sortPropertyMatches(matches []entity.PropertyMatch) []entity.PropertyMatch {
	// Separate into categories
	inquiredNotDismissed := make([]entity.PropertyMatch, 0)
	others := make([]entity.PropertyMatch, 0)

	for _, match := range matches {
		if match.IsInquired && !match.IsDismissed {
			inquiredNotDismissed = append(inquiredNotDismissed, match)
		} else if !match.IsDismissed {
			others = append(others, match)
		}
	}

	// Sort others by match percentage (descending)
	for i := 0; i < len(others)-1; i++ {
		for j := i + 1; j < len(others); j++ {
			if others[i].MatchPercent < others[j].MatchPercent {
				others[i], others[j] = others[j], others[i]
			}
		}
	}

	// Combine: inquired first, then others
	result := append(inquiredNotDismissed, others...)
	return result
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
