package cmd

import (
	"fmt"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var systemRmCmd = &cobra.Command{
	Use:   "rm [ref]",
	Short: "Remove a skill from the system store",
	Long: `Remove a skill artifact from the local system store.

This deletes the specified tag reference. The actual content (blobs) may remain
until 'skr system prune' is run, unless it is the last reference to that content.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		var errs []error
		for _, ref := range args {
			// Resolve first to ensure it exists
			desc, err := st.Resolve(ctx, ref)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					err := fmt.Errorf("reference %s not found in store", ref)
					fmt.Println(err)
					errs = append(errs, err)
					continue
				}
				err := fmt.Errorf("failed to resolve reference %s: %w", ref, err)
				fmt.Println(err)
				errs = append(errs, err)
				continue
			}

			fmt.Printf("Removing %s (%s)...\n", ref, desc.Digest)
			if err := st.Delete(ctx, desc); err != nil {
				err := fmt.Errorf("failed to remove artifact %s: %w", ref, err)
				fmt.Println(err)
				errs = append(errs, err)
				continue
			}
			fmt.Printf("Successfully removed %s\n", ref)
		}

		if len(errs) > 0 {
			return fmt.Errorf("failed to remove some artifacts")
		}
		return nil
	},
}

func init() {
	systemCmd.AddCommand(systemRmCmd)
}
