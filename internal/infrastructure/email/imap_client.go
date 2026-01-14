package email

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// EmailMessage represents a parsed email message
type EmailMessage struct {
	Subject string
	From    string
	Body    string
	UID     uint32
}

// IMAPClient handles IMAP connections and email retrieval
type IMAPClient struct {
	config Config
	client *client.Client
}

// NewIMAPClient creates a new IMAP client
func NewIMAPClient(config Config) (*IMAPClient, error) {
	return &IMAPClient{
		config: config,
	}, nil
}

// Connect establishes a connection to the IMAP server
func (c *IMAPClient) Connect() error {
	// Connect to server
	addr := fmt.Sprintf("%s:%d", c.config.IMAPHost, c.config.IMAPPort)

	log.Printf("[IMAP] Connecting to %s...", addr)

	var err error
	c.client, err = client.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	log.Printf("[IMAP] Connected to %s", addr)

	// Login
	if err := c.client.Login(c.config.Username, c.config.Password); err != nil {
		c.client.Logout()
		return fmt.Errorf("failed to login: %w", err)
	}

	log.Printf("[IMAP] Logged in as %s", c.config.Username)

	return nil
}

// Disconnect closes the IMAP connection
func (c *IMAPClient) Disconnect() error {
	if c.client != nil {
		return c.client.Logout()
	}
	return nil
}

// FetchUnreadEmails retrieves all unread emails from the inbox
func (c *IMAPClient) FetchUnreadEmails() ([]ParsedEmail, error) {
	// Connect first
	if err := c.Connect(); err != nil {
		return nil, err
	}

	// Select mailbox
	mbox, err := c.client.Select(c.config.InboxFolder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select mailbox: %w", err)
	}

	// No messages in mailbox
	if mbox.Messages == 0 {
		log.Printf("[IMAP] No messages in mailbox")
		return []ParsedEmail{}, nil
	}

	// Search for unread messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	uids, err := c.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search for unread messages: %w", err)
	}

	if len(uids) == 0 {
		log.Printf("[IMAP] No unread messages")
		return []ParsedEmail{}, nil
	}

	log.Printf("[IMAP] Found %d unread message(s)", len(uids))

	// Fetch messages
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, section.FetchItem()}

	go func() {
		done <- c.client.Fetch(seqset, items, messages)
	}()

	parsedEmails := []ParsedEmail{}

	for msg := range messages {
		if msg == nil {
			continue
		}

		parsedEmail := ParsedEmail{
			MessageID: fmt.Sprintf("%d", msg.Uid),
		}

		// Get envelope info
		if msg.Envelope != nil {
			parsedEmail.Subject = msg.Envelope.Subject
			if len(msg.Envelope.From) > 0 {
				parsedEmail.From = msg.Envelope.From[0].Address()
			}
			if !msg.Envelope.Date.IsZero() {
				parsedEmail.Date = msg.Envelope.Date
			}
		}

		// Get body
		r := msg.GetBody(section)
		if r != nil {
			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Printf("[IMAP] Error creating mail reader: %v", err)
				continue
			}

			// Read all parts
			body := ""
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("[IMAP] Error reading part: %v", err)
					break
				}

				switch h := p.Header.(type) {
				case *mail.InlineHeader:
					contentType, _, _ := h.ContentType()
					b, _ := io.ReadAll(p.Body)

					// Prefer HTML content, but also capture text
					if strings.HasPrefix(contentType, "text/html") {
						body = string(b)
						break // HTML is preferred, stop reading
					} else if strings.HasPrefix(contentType, "text/plain") && body == "" {
						body = string(b)
					}
				}
			}

			parsedEmail.Body = body
		}

		// Mark as read
		c.MarkAsRead(msg.Uid)

		parsedEmails = append(parsedEmails, parsedEmail)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return parsedEmails, nil
}

// MarkAsRead marks an email as read
func (c *IMAPClient) MarkAsRead(uid uint32) error {
	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}

	if err := c.client.Store(seqset, item, flags, nil); err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// Close closes the IMAP connection
func (c *IMAPClient) Close() error {
	return c.Disconnect()
}

// IsConnected checks if the client is connected
func (c *IMAPClient) IsConnected() bool {
	return c.client != nil && c.client.State() == imap.AuthenticatedState
}
