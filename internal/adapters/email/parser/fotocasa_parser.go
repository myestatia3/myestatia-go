package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"golang.org/x/net/html"
)

type FotocasaParser struct{}

func (p *FotocasaParser) CanParse(subject, from string) bool {
	from = strings.ToLower(from)
	subject = strings.ToLower(subject)

	return strings.Contains(from, "fotocasa") ||
		strings.Contains(subject, "fotocasa") ||
		strings.Contains(subject, "nuevo contacto de fotocasa")
}

func (p *FotocasaParser) Parse(subject, body string) (*entity.ParsedLead, error) {
	lead := &entity.ParsedLead{
		Source:      entity.EmailSourceFotocasa,
		ContactDate: time.Now(),
	}

	// Extract property reference from body or subject
	// Pattern: "referencia R4786633" or "con referencia R4786633"
	refRegex := regexp.MustCompile(`(?i)referencia\s+(R\d+)`)
	if matches := refRegex.FindStringSubmatch(body); len(matches) > 1 {
		lead.PropertyReference = matches[1]
	} else if matches := refRegex.FindStringSubmatch(subject); len(matches) > 1 {
		lead.PropertyReference = matches[1]
	}

	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return p.parseWithRegex(subject, body, lead)
	}

	extractedData := make(map[string]string)
	var textNodes []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textNodes = append(textNodes, text)
			}

			// Clean value helper to truncate at next label
			cleanValue := func(s string) string {
				s = strings.TrimSpace(s)
				if idx := strings.Index(s, "Nombre:"); idx != -1 {
					s = s[:idx]
				}
				if idx := strings.Index(s, "Teléfono:"); idx != -1 {
					s = s[:idx]
				}
				if idx := strings.Index(s, "Email:"); idx != -1 {
					s = s[:idx]
				}
				if idx := strings.Index(s, "Día y hora:"); idx != -1 {
					s = s[:idx]
				}
				if idx := strings.Index(s, "Mensaje:"); idx != -1 {
					s = s[:idx]
				}
				// Also truncate if we see "[image:" or "Referencia:" which might appear in the same block for message
				if idx := strings.Index(s, "[image:"); idx != -1 {
					s = s[:idx]
				}
				if idx := strings.Index(s, "Referencia:"); idx != -1 {
					s = s[:idx]
				}
				return strings.TrimSpace(s)
			}

			// Check for field labels and extract values - INDEPENDENT checks (no else)
			if idx := strings.Index(text, "Nombre:"); idx != -1 {
				val := text[idx+len("Nombre:"):]
				cleaned := cleanValue(val)
				if cleaned != "" && !strings.HasPrefix(strings.ToLower(cleaned), "llamar") {
					extractedData["name"] = cleaned
				} else {
					extractedData["nameLabel"] = text // Mark to maybe match next node, though less likely now
				}
			}

			if idx := strings.Index(text, "Teléfono:"); idx != -1 {
				val := text[idx+len("Teléfono:"):]
				cleaned := cleanValue(val)
				if cleaned != "" {
					extractedData["phone"] = cleaned
				} else {
					extractedData["phoneLabel"] = text
				}
			}

			if idx := strings.Index(text, "Email:"); idx != -1 {
				val := text[idx+len("Email:"):]
				cleaned := cleanValue(val)
				if cleaned != "" {
					extractedData["email"] = cleaned
				} else {
					extractedData["emailLabel"] = text
				}
			}

			if idx := strings.Index(text, "Mensaje:"); idx != -1 {
				val := text[idx+len("Mensaje:"):]
				// For message, we might be more lenient, but still cut off at unknown footer stuff
				cleaned := cleanValue(val)
				if cleaned != "" {
					extractedData["message"] = cleaned
				} else {
					extractedData["messageLabel"] = text
				}
			}

			// Handle case where label was in previous node (keeping existing logic for safety, though split logic above might cover it)
			if extractedData["nameLabel"] != "" && extractedData["name"] == "" && text != "" {
				if !strings.HasPrefix(strings.ToLower(text), "llamar") && !strings.Contains(text, ":") {
					extractedData["name"] = cleanValue(text)
					extractedData["nameLabel"] = ""
				}
			} else if extractedData["phoneLabel"] != "" && extractedData["phone"] == "" && text != "" {
				if !strings.Contains(text, ":") {
					extractedData["phone"] = cleanValue(text)
					extractedData["phoneLabel"] = ""
				}
			} // ... similar for others if needed but specific label matching above is stronger
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
		// Only update if not already found or if this seems more reliable
		foundPhone := strings.ReplaceAll(fullText[phoneLoc[0]:phoneLoc[1]], " ", "")
		extractedData["phone"] = foundPhone // prioritization

		// Search for email specifically after the phone number
		emailRegex := regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		if matches := emailRegex.FindStringSubmatch(fullText[phoneLoc[1]:]); len(matches) > 0 {
			extractedData["email"] = matches[0] // Override with specifically located email
		}
	}

	regexData := p.extractWithRegex(body)

	if extractedData["name"] != "" {
		lead.Name = extractedData["name"]
	} else if regexData["name"] != "" {
		lead.Name = regexData["name"]
	}

	if extractedData["phone"] != "" {
		lead.Phone = extractedData["phone"]
	} else if regexData["phone"] != "" {
		lead.Phone = regexData["phone"]
	}

	if extractedData["email"] != "" {
		lead.Email = extractedData["email"]
	} else if regexData["email"] != "" {
		lead.Email = regexData["email"]
	}

	if extractedData["message"] != "" {
		lead.Message = extractedData["message"]
	} else if regexData["message"] != "" {
		lead.Message = regexData["message"]
	}

	// Validate required fields
	if lead.PropertyReference == "" {
		return nil, fmt.Errorf("could not extract property reference from Fotocasa email")
	}
	if lead.Email == "" {
		return nil, fmt.Errorf("could not extract email from Fotocasa email")
	}

	return lead, nil
}

// extractWithRegex extracts data using regex patterns
func (p *FotocasaParser) extractWithRegex(body string) map[string]string {
	data := make(map[string]string)

	// Extract name: "Nombre: <value>"
	nameRegex := regexp.MustCompile(`(?i)Nombre:\s*([^\n<]+)`)
	if matches := nameRegex.FindStringSubmatch(body); len(matches) > 1 {
		data["name"] = strings.TrimSpace(matches[1])
	}

	// Extract phone: "Teléfono: <value>"
	phoneRegex := regexp.MustCompile(`(?i)Tel[eé]fono:\s*([0-9\s]+)`)
	if matches := phoneRegex.FindStringSubmatch(body); len(matches) > 1 {
		data["phone"] = strings.TrimSpace(matches[1])
	}

	// Extract email: "Email: <value>"
	emailRegex := regexp.MustCompile(`(?i)Email:\s*([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
	if matches := emailRegex.FindStringSubmatch(body); len(matches) > 1 {
		data["email"] = strings.TrimSpace(matches[1])
	}

	// Extract message: "Mensaje: <value>"
	messageRegex := regexp.MustCompile(`(?i)Mensaje:\s*([^\n<]+(?:\n[^\n<]+)*)`)
	if matches := messageRegex.FindStringSubmatch(body); len(matches) > 1 {
		data["message"] = strings.TrimSpace(matches[1])
	}

	return data
}

// parseWithRegex is a complete regex-based fallback parser
func (p *FotocasaParser) parseWithRegex(subject, body string, lead *entity.ParsedLead) (*entity.ParsedLead, error) {
	regexData := p.extractWithRegex(body)

	if regexData["name"] != "" {
		lead.Name = regexData["name"]
	}
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
		return nil, fmt.Errorf("could not extract property reference from Fotocasa email")
	}
	if lead.Email == "" {
		return nil, fmt.Errorf("could not extract email from Fotocasa email")
	}

	return lead, nil
}
