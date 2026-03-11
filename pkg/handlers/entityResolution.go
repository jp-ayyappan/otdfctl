package handlers

import (
	"context"

	"github.com/opentdf/platform/protocol/go/entity"
	ersv2 "github.com/opentdf/platform/protocol/go/entityresolution/v2"
)

// ResolveEntities calls the Entity Resolution Service to return the attribute
// entitlements for the given entities.
//
// The SDK connects to the ERS endpoint discovered from the platform well-known
// configuration. If ERS is at a different host, pass the ersEndpoint hint for
// informational purposes — a future enhancement will support custom endpoints.
func (h Handler) ResolveEntities(ctx context.Context, entities []*entity.Entity) (*ersv2.ResolveEntitiesResponse, error) {
	req := &ersv2.ResolveEntitiesRequest{Entities: entities}
	return h.sdk.EntityResolutionV2.ResolveEntities(ctx, req)
}
