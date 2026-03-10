package tui

import (
	"context"

	"github.com/opentdf/platform/protocol/go/common"
	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/protocol/go/policy/attributes"
	"github.com/opentdf/platform/protocol/go/policy/kasregistry"
	"github.com/opentdf/platform/protocol/go/policy/namespaces"
	"github.com/opentdf/platform/protocol/go/policy/resourcemapping"
	"github.com/opentdf/platform/protocol/go/policy/subjectmapping"

	"github.com/opentdf/otdfctl/pkg/handlers"
)

// TUIHandler is the subset of handlers.Handler used by TUI views.
// Defining an interface here allows views to be unit-tested with a mock
// without requiring a live OpenTDF platform connection.
//
// *handlers.Handler satisfies this interface (pointer needed for methods
// that have pointer receivers: UpdateAttribute, GetResourceMapping,
// ListResourceMappings).
type TUIHandler interface {
	GetEndpoint() string
	GetProfileName() string

	// Attributes
	ListAttributes(ctx context.Context, state common.ActiveStateEnum, limit, offset int32) (*attributes.ListAttributesResponse, error)
	GetAttribute(ctx context.Context, identifier string) (*policy.Attribute, error)
	UpdateAttribute(ctx context.Context, id string, metadata *common.MetadataMutable, behavior common.MetadataUpdateEnum) (*policy.Attribute, error)

	// Namespaces
	ListNamespaces(ctx context.Context, state common.ActiveStateEnum, limit, offset int32) (*namespaces.ListNamespacesResponse, error)
	GetNamespace(ctx context.Context, identifier string) (*policy.Namespace, error)

	// Subject Mappings
	ListSubjectMappings(ctx context.Context, limit, offset int32) (*subjectmapping.ListSubjectMappingsResponse, error)
	GetSubjectMapping(ctx context.Context, id string) (*policy.SubjectMapping, error)

	// KAS Registry
	ListKasRegistryEntries(ctx context.Context, limit, offset int32) (*kasregistry.ListKeyAccessServersResponse, error)
	GetKasRegistryEntry(ctx context.Context, identifier handlers.KasIdentifier) (*policy.KeyAccessServer, error)

	// Resource Mappings
	ListResourceMappings(ctx context.Context, limit, offset int32) (*resourcemapping.ListResourceMappingsResponse, error)
	GetResourceMapping(id string) (*policy.ResourceMapping, error)
}
