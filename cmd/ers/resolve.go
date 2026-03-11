package ers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/opentdf/otdfctl/cmd/common"
	"github.com/opentdf/otdfctl/pkg/cli"
	"github.com/opentdf/platform/protocol/go/entity"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve attribute entitlements for one or more entities",
	Long: `Resolve attribute entitlements by querying the Entity Resolution Service.

The response contains two sections per entity:
  - Identity attributes  : raw properties returned by the identity provider
                           (Keycloak/LDAP fields such as clearance, nationality)
  - OpenTDF entitlements : computed attribute-value FQNs from subject mappings
                           (empty when no subject mappings match the entity)

Examples:
  otdfctl ers resolve --email alice@example.com
  otdfctl ers resolve --email alice@example.com --email bob@example.com
  otdfctl ers resolve --client-id my-service-account
  otdfctl ers resolve --email alice@example.com --ers-endpoint https://ers.internal.example.com`,
	RunE: runResolve,
}

var (
	flagEmails      []string
	flagClientIDs   []string
	flagUsernames   []string
	flagERSEndpoint string
)

func init() {
	resolveCmd.Flags().StringArrayVar(&flagEmails, "email", nil, "Email address of entity to resolve (repeatable)")
	resolveCmd.Flags().StringArrayVar(&flagClientIDs, "client-id", nil, "Client ID of entity to resolve (repeatable)")
	resolveCmd.Flags().StringArrayVar(&flagUsernames, "username", nil, "Username of entity to resolve (repeatable)")
	resolveCmd.Flags().StringVar(&flagERSEndpoint, "ers-endpoint", "", "ERS endpoint override (default: same host as platform)")
}

func runResolve(cmd *cobra.Command, args []string) error {
	c := cli.New(cmd, args)
	h := common.NewHandler(c)
	defer h.Close()

	if len(flagEmails)+len(flagClientIDs)+len(flagUsernames) == 0 {
		return fmt.Errorf("at least one of --email, --client-id, or --username is required")
	}

	if flagERSEndpoint != "" {
		h.SetERSEndpoint(flagERSEndpoint)
	}

	var entities []*entity.Entity
	for _, e := range flagEmails {
		entities = append(entities, &entity.Entity{EntityType: &entity.Entity_EmailAddress{EmailAddress: e}})
	}
	for _, cid := range flagClientIDs {
		entities = append(entities, &entity.Entity{EntityType: &entity.Entity_ClientId{ClientId: cid}})
	}
	for _, u := range flagUsernames {
		entities = append(entities, &entity.Entity{EntityType: &entity.Entity_UserName{UserName: u}})
	}

	resp, err := h.ResolveEntities(cmd.Context(), entities)
	if err != nil {
		cli.ExitWithError("Failed to resolve entities via ERS", err)
	}

	reps := resp.GetEntityRepresentations()
	if len(reps) == 0 {
		fmt.Println("No entity representations returned.")
		return nil
	}

	for i, rep := range reps {
		// Map the ephemeral index back to the original input identifier
		label := ersLabel(rep.GetOriginalId(), entities, i)
		fmt.Printf("\n── Entity: %s ──\n", label)

		// ── Identity attributes from the identity provider ──
		idRows, attrRows := extractProps(rep.GetAdditionalProps())
		if len(idRows) > 0 {
			fmt.Println("\nIdentity:")
			t := cli.NewTabular(idRows...)
			cli.PrintSuccessTable(cmd, "", t)
		}
		if len(attrRows) > 0 {
			fmt.Println("\nIdentity Provider Attributes:")
			t := cli.NewTabular(attrRows...)
			cli.PrintSuccessTable(cmd, "", t)
		}

		// ── OpenTDF policy entitlements ──
		entitlements := rep.GetDirectEntitlements()
		if len(entitlements) == 0 {
			fmt.Println("\nOpenTDF Entitlements: (none — check subject mappings)")
		} else {
			fmt.Println("\nOpenTDF Entitlements:")
			rows := make([][]string, 0, len(entitlements))
			for _, e := range entitlements {
				rows = append(rows, []string{e.GetAttributeValueFqn(), strings.Join(e.GetActions(), ", ")})
			}
			t := cli.NewTabular(rows...)
			cli.PrintSuccessTable(cmd, "", t)
		}
	}
	return nil
}

// ersLabel maps the ephemeral "entity_idx_N" back to the original identifier.
func ersLabel(originalID string, entities []*entity.Entity, idx int) string {
	if idx < len(entities) {
		e := entities[idx]
		switch et := e.GetEntityType().(type) {
		case *entity.Entity_EmailAddress:
			return et.EmailAddress
		case *entity.Entity_ClientId:
			return et.ClientId
		case *entity.Entity_UserName:
			return et.UserName
		}
	}
	return originalID
}

// internalLDAPKeys are Keycloak/LDAP bookkeeping fields we skip in output.
var internalLDAPKeys = map[string]bool{
	"LDAP_ENTRY_DN":   true,
	"LDAP_ID":         true,
	"createTimestamp": true,
	"modifyTimestamp": true,
}

// extractProps pulls identity fields and custom attributes from additional_props.
// Returns two row slices: identity fields and custom attribute key/value rows.
func extractProps(props []*structpb.Struct) (idRows, attrRows [][]string) {
	identityKeys := []string{"email", "username", "firstName", "lastName", "id"}

	for _, prop := range props {
		if prop == nil {
			continue
		}
		fields := prop.GetFields()

		// Identity fields
		for _, key := range identityKeys {
			if v, ok := fields[key]; ok && v.GetStringValue() != "" {
				idRows = append(idRows, []string{key, v.GetStringValue()})
			}
		}

		// Custom attributes nested under "attributes"
		if attrField, ok := fields["attributes"]; ok {
			if attrStruct := attrField.GetStructValue(); attrStruct != nil {
				keys := make([]string, 0, len(attrStruct.GetFields()))
				for k := range attrStruct.GetFields() {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					if internalLDAPKeys[k] {
						continue
					}
					var vals []string
					for _, item := range attrStruct.GetFields()[k].GetListValue().GetValues() {
						vals = append(vals, item.GetStringValue())
					}
					if len(vals) > 0 {
						attrRows = append(attrRows, []string{k, strings.Join(vals, ", ")})
					}
				}
			}
		}
	}
	return
}

