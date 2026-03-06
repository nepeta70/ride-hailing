package service

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type GrainIdentityFactory struct{}

var rideGrainNamespace = uuid.MustParse("6ba7b810-9dad-78d1-80b4-00c04fd430c8")

func (f *GrainIdentityFactory) NewRideGrainIdentity(userID uuid.UUID, requestID uuid.UUID, fareID uuid.UUID) *grain.GrainIdentity {
	input := userID.String() + ":" + requestID.String() + ":" + fareID.String()
	identity := &grain.GrainIdentity{
		Kind:     domain.RideGrainKind,
		EntityID: uuid.NewSHA1(rideGrainNamespace, []byte(input)),
	}
	return identity
}
