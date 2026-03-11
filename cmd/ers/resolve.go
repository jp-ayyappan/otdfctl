package ers

import (
	"fmt"
	"strings"

	"github.com/opentdf/otdfctl/cmd/common"
	"github.com/opentdf/otdfctl/pkg/cli"
	"github.com/opentdf/platform/protocol/go/entity"
	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve attribute entitlements for one or more entities",
	Long: `Resolve attribute entitlements by querying the Entity Resolution Service.

Provide at least one entity identifier. Multiple flags may be combined to
resolve several entities in a single request.

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

	for _, rep := range reps {
		entitlements := rep.GetDirectEntitlements()
		rows := make([][]string, 0, len(entitlements))
		for _, e := range entitlements {
			rows = append(rows, []string{
				e.GetAttributeValueFqn(),
				strings.Join(e.GetActions(), ", "),
			})
		}
		t := cli.NewTabular(rows...)
		cli.PrintSuccessTable(cmd, rep.GetOriginalId(), t)
	}
	return nil
}
