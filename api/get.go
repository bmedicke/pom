package pom

import (
	"context"
	"log"

	get "github.com/bmedicke/pom/api/gen/get"
)

// get service example implementation.
// The example methods log the requests and return zero values.
type getsrvc struct {
	logger *log.Logger
}

// NewGet returns the get service implementation.
func NewGet(logger *log.Logger) get.Service {
	return &getsrvc{logger}
}

// State implements state.
func (s *getsrvc) State(ctx context.Context) (res string, err error) {
	s.logger.Print("get.state")
	return
}
