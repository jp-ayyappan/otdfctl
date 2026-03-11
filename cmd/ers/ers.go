// Package ers provides the "ers" top-level command for interacting with the
// Entity Resolution Service.
package ers

import "github.com/spf13/cobra"

// Cmd is the root "ers" command.
var Cmd = &cobra.Command{
	Use:   "ers",
	Short: "Entity Resolution Service — resolve entity attributes",
	Long: `Query the Entity Resolution Service (ERS) to look up the attribute
entitlements associated with one or more entities (users or clients).

By default ERS is assumed to be co-located with the platform endpoint
stored in the active profile. Use --ers-endpoint to override.`,
}

func init() {
	Cmd.AddCommand(resolveCmd)
}
