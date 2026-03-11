package handlers

import (
	"errors"
	"log/slog"

	"github.com/opentdf/otdfctl/pkg/auth"
	"github.com/opentdf/otdfctl/pkg/profiles"
	"github.com/opentdf/otdfctl/pkg/utils"
	"github.com/opentdf/platform/protocol/go/common"
	"github.com/opentdf/platform/sdk"
)

var (
	SDK *sdk.SDK

	ErrUnauthenticated = errors.New("unauthenticated")
)

type Handler struct {
	sdk              *sdk.SDK
	platformEndpoint string
	profileName      string
	ersEndpoint      string // optional ERS endpoint override; defaults to platformEndpoint
}

// SetERSEndpoint overrides the ERS endpoint used for entity resolution.
// Defaults to platformEndpoint if not set.
func (h *Handler) SetERSEndpoint(endpoint string) {
	h.ersEndpoint = endpoint
}

// GetERSEndpoint returns the active ERS endpoint (override or platform fallback).
func (h Handler) GetERSEndpoint() string {
	if h.ersEndpoint != "" {
		return h.ersEndpoint
	}
	return h.platformEndpoint
}

type handlerOpts struct {
	endpoint    string
	TLSNoVerify bool

	profile *profiles.OtdfctlProfileStore

	sdkOpts []sdk.Option
}

type handlerOptsFunc func(handlerOpts) handlerOpts

func WithEndpoint(endpoint string, tlsNoVerify bool) handlerOptsFunc {
	return func(c handlerOpts) handlerOpts {
		c.endpoint = endpoint
		c.TLSNoVerify = tlsNoVerify
		return c
	}
}

func WithProfile(profile *profiles.OtdfctlProfileStore) handlerOptsFunc {
	return func(c handlerOpts) handlerOpts {
		c.profile = profile
		c.endpoint = profile.GetEndpoint()
		c.TLSNoVerify = profile.GetTLSNoVerify()

		// get sdk opts
		opts, err := auth.GetSDKAuthOptionFromProfile(profile)
		if err != nil {
			return c
		}
		c.sdkOpts = append(c.sdkOpts, opts)

		return c
	}
}

func WithSDKOpts(opts ...sdk.Option) handlerOptsFunc {
	return func(c handlerOpts) handlerOpts {
		c.sdkOpts = opts
		return c
	}
}

// Creates a new handler wrapping the SDK, which is authenticated through the cached client-credentials flow tokens
func New(opts ...handlerOptsFunc) (Handler, error) {
	var o handlerOpts
	for _, f := range opts {
		o = f(o)
	}

	u, err := utils.NormalizeEndpoint(o.endpoint)
	if err != nil {
		return Handler{}, err
	}

	// get auth
	authSDKOpt, err := auth.GetSDKAuthOptionFromProfile(o.profile)
	if err != nil {
		return Handler{}, err
	}

	defaultSDKOpts := []sdk.Option{
		authSDKOpt,
		sdk.WithConnectionValidation(),
		sdk.WithLogger(slog.Default()),
	}
	if o.TLSNoVerify {
		defaultSDKOpts = append(defaultSDKOpts, sdk.WithInsecureSkipVerifyConn())
	}

	if u.Scheme == "http" {
		defaultSDKOpts = append(defaultSDKOpts, sdk.WithInsecurePlaintextConn())
	}
	o.sdkOpts = append(defaultSDKOpts, o.sdkOpts...)

	s, err := sdk.New(u.String(), o.sdkOpts...)
	if err != nil {
		return Handler{}, err
	}

	profileName := ""
	if o.profile != nil {
		profileName = o.profile.Name()
	}

	return Handler{
		sdk:              s,
		platformEndpoint: o.endpoint,
		profileName:      profileName,
	}, nil
}

func (h Handler) Close() error {
	return h.sdk.Close()
}

func (h Handler) Direct() *sdk.SDK {
	return h.sdk
}

func (h Handler) GetEndpoint() string {
	return h.platformEndpoint
}

func (h Handler) GetProfileName() string {
	return h.profileName
}

// Replace all labels in the metadata
func (h Handler) WithReplaceLabelsMetadata(metadata *common.MetadataMutable, labels map[string]string) func(*common.MetadataMutable) *common.MetadataMutable {
	return func(*common.MetadataMutable) *common.MetadataMutable {
		nextMetadata := &common.MetadataMutable{
			Labels: labels,
		}
		return nextMetadata
	}
}

// Append a label to the metadata
func (h Handler) WithLabelMetadata(metadata *common.MetadataMutable, key, value string) func(*common.MetadataMutable) *common.MetadataMutable {
	return func(*common.MetadataMutable) *common.MetadataMutable {
		labels := metadata.GetLabels()
		labels[key] = value
		nextMetadata := &common.MetadataMutable{
			Labels: labels,
		}
		return nextMetadata
	}
}

// func buildMetadata(metadata *common.MetadataMutable, fns ...func(*common.MetadataMutable) *common.MetadataMutable) *common.MetadataMutable {
// 	if metadata == nil {
// 		metadata = &common.MetadataMutable{}
// 	}
// 	if len(fns) == 0 {
// 		return metadata
// 	}
// 	for _, fn := range fns {
// 		metadata = fn(metadata)
// 	}
// 	return metadata
// }
