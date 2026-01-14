package parser

import (
	"fmt"

	"github.com/myestatia/myestatia-go/internal/domain/port"
)

type ParserFactory struct {
	parsers []port.EmailParser
}

func NewParserFactory() *ParserFactory {
	factory := &ParserFactory{
		parsers: make([]port.EmailParser, 0),
	}

	factory.RegisterParser(&FotocasaParser{})
	factory.RegisterParser(&IdealistaParser{})

	return factory
}

func (f *ParserFactory) RegisterParser(parser port.EmailParser) {
	f.parsers = append(f.parsers, parser)
}

func (f *ParserFactory) GetParser(subject, from string) (port.EmailParser, error) {
	for _, parser := range f.parsers {
		if parser.CanParse(subject, from) {
			return parser, nil
		}
	}

	return nil, fmt.Errorf("no suitable parser found for email from '%s' with subject '%s'", from, subject)
}
