package cmdg

import (
	"context"
	"sync"

	gmail "google.golang.org/api/gmail/v1"
)

// Page implements some pagination thingy. TODO: document better.
type Page struct {
	Label string
	Query string

	m sync.RWMutex

	conn     *CmdG
	Messages []*Message
	Response *gmail.ListMessagesResponse
}

// Next return the next page.
func (p *Page) Next(ctx context.Context) (*Page, error) {
	return p.conn.ListMessages(ctx, p.Label, p.Query, p.Response.NextPageToken)
}

// PreloadSubjects async loads message basic info.
func (p *Page) PreloadSubjects(ctx context.Context) error {
	conc := 100
	sem := make(chan struct{}, conc)
	num := len(p.Response.Messages)
	errs := make([]error, num, num)
	for n := 0; n < len(p.Response.Messages); n++ {
		n := n
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()

			if err := p.Messages[n].Preload(ctx, LevelMetadata); err != nil {
				errs[n] = err
			}
		}()
	}
	for t := 0; t < conc; t++ {
		sem <- struct{}{}
	}
	return nil
}
