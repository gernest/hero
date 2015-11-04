package hero

import (
	"github.com/pborman/uuid"
)

// SimpleTokenGen implements TokenGenerator interface
type SimpleTokenGen struct{}

// Generate returns a UUID v4 string
func (s *SimpleTokenGen) Generate() string {
	return uuid.NewRandom().String()
}
