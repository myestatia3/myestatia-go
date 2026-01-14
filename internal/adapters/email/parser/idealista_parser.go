package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"golang.org/x/net/html"
)

type IdealistaParser struct{}

// CanParse determines if this email is from Idealista
func (p *IdealistaParser) CanParse(subject, from string) bool {
	from = strings.ToLower(from)
	subject = strings.ToLower(subject)

	// Check if from Idealista domain or subject contains Idealista patterns
	return strings.Contains(from, "idealista") ||
		strings.Contains(subject, "idealista") ||
		strings.Contains(subject, "nuevo mensaje")
}

// Parse extracts lead data from Idealista email HTML
func (p *IdealistaParser) Parse(subject, body string) (*entity.ParsedLead, error) {
	lead := &entity.ParsedLead{
		Source:      entity.EmailSourceIdealista,
		ContactDate: time.Now(),
	}

	// Extract property reference from body or subject
	// Pattern: "Ref. R4967962" or "ref: R4967962"
	refRegex := regexp.MustCompile(`(?i)ref[.:]?\s*(R\d+)`)
	if matches := refRegex.FindStringSubmatch(body); len(matches) > 1 {
		lead.PropertyReference = matches[1]
	} else if matches := refRegex.FindStringSubmatch(subject); len(matches) > 1 {
		lead.PropertyReference = matches[1]
	}

	// Extract name from subject: "Nuevo mensaje de [Nombre] sobre..."
	nameInSubjectRegex := regexp.MustCompile(`(?i)mensaje de\s+([^s]+?)\s+sobre`)
	if matches := nameInSubjectRegex.FindStringSubmatch(subject); len(matches) > 1 {
		lead.Name = strings.TrimSpace(matches[1])
	}

	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return p.parseWithRegex(subject, body, lead)
	}

	var textNodes []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textNodes = append(textNodes, text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	fullText := strings.Join(textNodes, " ")

	phoneRegex := regexp.MustCompile(`(\d{3}\s*\d{2}\s*\d{2}\s*\d{2}|\d{9})`)
	phoneLoc := phoneRegex.FindStringIndex(fullText)
	if phoneLoc != nil {
		lead.Phone = strings.ReplaceAll(fullText[phoneLoc[0]:phoneLoc[1]], " ", "")

		// Search for email specifically after the phone number
		// This avoids picking up the "From" or "To" emails in the header
		emailRegex := regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		if matches := emailRegex.FindStringSubmatch(fullText[phoneLoc[1]:]); len(matches) > 0 {
			lead.Email = matches[0]
		}
	}

	// Fallback/Original logic if phone not found or email not found after phone
	if lead.Email == "" {
		emailRegex := regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		if matches := emailRegex.FindStringSubmatch(fullText); len(matches) > 0 {
			lead.Email = matches[0]
		}
	}

	// Extract message - look for the main message content
	// Usually appears after contact info
	messageRegex := regexp.MustCompile(`(?i)(Hola[^.]+(?:\.[^.]+){0,5}|me gustar[ií]a[^.]+(?:\.[^.]+){0,5}|estoy interesad[oa][^.]+(?:\.[^.]+){0,5})`)
	if matches := messageRegex.FindStringSubmatch(fullText); len(matches) > 0 {
		lead.Message = strings.TrimSpace(matches[0])
	}

	// Also try regex extraction as fallback
	regexData := p.extractWithRegex(body)

	// Merge data, preferring HTML parsed data but filling in gaps
	if lead.Name == "" && regexData["name"] != "" {
		lead.Name = regexData["name"]
	}
	if lead.Phone == "" && regexData["phone"] != "" {
		lead.Phone = regexData["phone"]
	}
	if lead.Email == "" && regexData["email"] != "" {
		lead.Email = regexData["email"]
	}
	if lead.Message == "" && regexData["message"] != "" {
		lead.Message = regexData["message"]
	}

	// Validate required fields
	if lead.PropertyReference == "" {
		return nil, fmt.Errorf("could not extract property reference from Idealista email")
	}
	if lead.Email == "" {
		return nil, fmt.Errorf("could not extract email from Idealista email")
	}

	return lead, nil
}

// extractWithRegex extracts data using regex patterns
func (p *IdealistaParser) extractWithRegex(body string) map[string]string {
	data := make(map[string]string)

	// Extract phone number
	phoneRegex := regexp.MustCompile(`(\d{3}\s*\d{2}\s*\d{2}\s*\d{2}|\d{9})`)
	if matches := phoneRegex.FindStringSubmatch(body); len(matches) > 0 {
		data["phone"] = strings.ReplaceAll(matches[0], " ", "")
	}

	// Extract email
	emailRegex := regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
	if matches := emailRegex.FindStringSubmatch(body); len(matches) > 0 {
		data["email"] = matches[0]
	}

	// Extract message content
	messageRegex := regexp.MustCompile(`(?i)(Hola[^<]+|me gustar[ií]a[^<]+|estoy interesad[oa][^<]+)`)
	if matches := messageRegex.FindStringSubmatch(body); len(matches) > 0 {
		data["message"] = strings.TrimSpace(matches[0])
	}

	return data
}

// parseWithRegex is a complete regex-based fallback parser
func (p *IdealistaParser) parseWithRegex(subject, body string, lead *entity.ParsedLead) (*entity.ParsedLead, error) {
	regexData := p.extractWithRegex(body)

	if regexData["phone"] != "" {
		lead.Phone = regexData["phone"]
	}
	if regexData["email"] != "" {
		lead.Email = regexData["email"]
	}
	if regexData["message"] != "" {
		lead.Message = regexData["message"]
	}

	if lead.PropertyReference == "" {
		return nil, fmt.Errorf("could not extract property reference from Idealista email")
	}
	if lead.Email == "" {
		return nil, fmt.Errorf("could not extract email from Idealista email")
	}

	return lead, nil
}
