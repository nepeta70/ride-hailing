package grain

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
)

// GrainIdentity uniquely identifies a grain instance
// Composed of kind (grain type) and entityID (business ID)
type GrainIdentity struct {
	Kind     enums.AggregateType
	EntityID uuid.UUID
}

// NewGrainIdentity creates a properly formatted grain identity
func NewGrainIdentity(kind enums.AggregateType, entityID uuid.UUID) *GrainIdentity {
	return &GrainIdentity{
		Kind:     kind,
		EntityID: entityID,
	}
}

// String returns the composite key "kind:entityID"
func (g *GrainIdentity) String() string {
	return g.Kind.String() + ":" + g.EntityID.String()
}
