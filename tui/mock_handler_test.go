package tui

import (
	"context"
	"errors"

	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/platform/protocol/go/common"
	"github.com/opentdf/platform/protocol/go/entity"
	ersv2 "github.com/opentdf/platform/protocol/go/entityresolution/v2"
	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/protocol/go/policy/attributes"
	"github.com/opentdf/platform/protocol/go/policy/kasregistry"
	"github.com/opentdf/platform/protocol/go/policy/namespaces"
	"github.com/opentdf/platform/protocol/go/policy/resourcemapping"
	"github.com/opentdf/platform/protocol/go/policy/subjectmapping"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// mockHandler is a test double that satisfies TUIHandler.
// Fields control what each method returns; defaults return empty non-nil responses.
type mockHandler struct {
	endpoint    string
	profileName string

	attributes      []*policy.Attribute
	attributeErr    error
	namespaces      []*policy.Namespace
	namespaceErr    error
	subjectMappings []*policy.SubjectMapping
	smErr           error
	kasEntries      []*policy.KeyAccessServer
	kasErr          error
	resourceMaps    []*policy.ResourceMapping
	rmErr           error
}

// Compile-time assertion that mockHandler satisfies TUIHandler.
var _ TUIHandler = (*mockHandler)(nil)

func newMock() *mockHandler { return &mockHandler{endpoint: "https://test.example.com", profileName: "test-profile"} }

func (m *mockHandler) GetEndpoint() string    { return m.endpoint }
func (m *mockHandler) GetProfileName() string { return m.profileName }

func (m *mockHandler) ListAttributes(_ context.Context, _ common.ActiveStateEnum, _, _ int32) (*attributes.ListAttributesResponse, error) {
	if m.attributeErr != nil {
		return nil, m.attributeErr
	}
	return &attributes.ListAttributesResponse{Attributes: m.attributes}, nil
}

func (m *mockHandler) GetAttribute(_ context.Context, id string) (*policy.Attribute, error) {
	if m.attributeErr != nil {
		return nil, m.attributeErr
	}
	for _, a := range m.attributes {
		if a.GetId() == id {
			return a, nil
		}
	}
	return &policy.Attribute{Id: id, Name: "test-attr"}, nil
}

func (m *mockHandler) UpdateAttribute(_ context.Context, id string, _ *common.MetadataMutable, _ common.MetadataUpdateEnum) (*policy.Attribute, error) {
	return &policy.Attribute{Id: id}, m.attributeErr
}

func (m *mockHandler) ListNamespaces(_ context.Context, _ common.ActiveStateEnum, _, _ int32) (*namespaces.ListNamespacesResponse, error) {
	if m.namespaceErr != nil {
		return nil, m.namespaceErr
	}
	return &namespaces.ListNamespacesResponse{Namespaces: m.namespaces}, nil
}

func (m *mockHandler) GetNamespace(_ context.Context, id string) (*policy.Namespace, error) {
	if m.namespaceErr != nil {
		return nil, m.namespaceErr
	}
	for _, ns := range m.namespaces {
		if ns.GetId() == id {
			return ns, nil
		}
	}
	return &policy.Namespace{Id: id, Name: "test-ns", Active: wrapperspb.Bool(true)}, nil
}

func (m *mockHandler) ListSubjectMappings(_ context.Context, _, _ int32) (*subjectmapping.ListSubjectMappingsResponse, error) {
	if m.smErr != nil {
		return nil, m.smErr
	}
	return &subjectmapping.ListSubjectMappingsResponse{SubjectMappings: m.subjectMappings}, nil
}

func (m *mockHandler) GetSubjectMapping(_ context.Context, id string) (*policy.SubjectMapping, error) {
	if m.smErr != nil {
		return nil, m.smErr
	}
	for _, sm := range m.subjectMappings {
		if sm.GetId() == id {
			return sm, nil
		}
	}
	return &policy.SubjectMapping{Id: id}, nil
}

func (m *mockHandler) ListSubjectConditionSets(_ context.Context, _, _ int32) (*subjectmapping.ListSubjectConditionSetsResponse, error) {
	return &subjectmapping.ListSubjectConditionSetsResponse{}, m.smErr
}

func (m *mockHandler) GetSubjectConditionSet(_ context.Context, id string) (*policy.SubjectConditionSet, error) {
	return &policy.SubjectConditionSet{Id: id}, m.smErr
}

func (m *mockHandler) ListKasRegistryEntries(_ context.Context, _, _ int32) (*kasregistry.ListKeyAccessServersResponse, error) {
	if m.kasErr != nil {
		return nil, m.kasErr
	}
	return &kasregistry.ListKeyAccessServersResponse{KeyAccessServers: m.kasEntries}, nil
}

func (m *mockHandler) GetKasRegistryEntry(_ context.Context, id handlers.KasIdentifier) (*policy.KeyAccessServer, error) {
	if m.kasErr != nil {
		return nil, m.kasErr
	}
	for _, k := range m.kasEntries {
		if k.GetId() == id.ID {
			return k, nil
		}
	}
	return &policy.KeyAccessServer{Id: id.ID, Uri: "https://kas.example.com"}, nil
}

func (m *mockHandler) ListAttributeValues(_ context.Context, _ string) ([]*policy.Value, error) {
	return nil, m.attributeErr
}

func (m *mockHandler) GetAttributeValue(_ context.Context, id string) (*policy.Value, error) {
	return &policy.Value{Id: id, Value: "test-value", Fqn: "https://ns.io/attr/a/value/test-value"}, m.attributeErr
}

func (m *mockHandler) ListKasKeys(_ context.Context, _, _ int32, _ policy.Algorithm, _ handlers.KasIdentifier, _ *bool) (*kasregistry.ListKeysResponse, error) {
	return &kasregistry.ListKeysResponse{}, m.kasErr
}

func (m *mockHandler) ListKeyMappings(_ context.Context, _, _ int32, _ string, _ *kasregistry.KasKeyIdentifier) (*kasregistry.ListKeyMappingsResponse, error) {
	return &kasregistry.ListKeyMappingsResponse{}, m.kasErr
}

func (m *mockHandler) ListResourceMappings(_ context.Context, _, _ int32) (*resourcemapping.ListResourceMappingsResponse, error) {
	if m.rmErr != nil {
		return nil, m.rmErr
	}
	return &resourcemapping.ListResourceMappingsResponse{ResourceMappings: m.resourceMaps}, nil
}

func (m *mockHandler) GetResourceMapping(id string) (*policy.ResourceMapping, error) {
	if m.rmErr != nil {
		return nil, m.rmErr
	}
	for _, rm := range m.resourceMaps {
		if rm.GetId() == id {
			return rm, nil
		}
	}
	return &policy.ResourceMapping{Id: id, Terms: []string{"term1", "term2"}}, nil
}

func (m *mockHandler) ResolveEntities(_ context.Context, _ []*entity.Entity) (*ersv2.ResolveEntitiesResponse, error) {
	return &ersv2.ResolveEntitiesResponse{}, m.kasErr
}

// ---- helpers for building test proto objects ----

var errTest = errors.New("simulated API error")

func testAttr(id, name string) *policy.Attribute {
	return &policy.Attribute{
		Id:   id,
		Name: name,
		Metadata: &common.Metadata{
			Labels: map[string]string{"env": "test"},
		},
	}
}

func testNamespace(id, name string) *policy.Namespace {
	return &policy.Namespace{Id: id, Name: name, Active: wrapperspb.Bool(true)}
}

func testSubjectMapping(id, fqn string) *policy.SubjectMapping {
	return &policy.SubjectMapping{
		Id: id,
		AttributeValue: &policy.Value{Fqn: fqn},
		Actions:        []*policy.Action{{Name: "read"}},
	}
}

func testKAS(id, name, uri string) *policy.KeyAccessServer {
	return &policy.KeyAccessServer{Id: id, Name: name, Uri: uri}
}

func testResourceMapping(id string, terms []string) *policy.ResourceMapping {
	return &policy.ResourceMapping{Id: id, Terms: terms}
}

// stdWindowMsg returns a typical terminal window size message.
func stdWindowMsg() interface{ String() string } {
	return nil // placeholder — tests import tea directly
}
