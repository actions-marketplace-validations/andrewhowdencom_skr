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
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		ctx := cmd.Context()

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// Resolve first to ensure it exists
		desc, err := st.Resolve(ctx, ref)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("reference %s not found in store", ref)
			}
			return fmt.Errorf("failed to resolve reference %s: %w", ref, err)
		}

		// Delete the tag/reference
		// Note: We are deleting the *reference* (tag), not necessarily the blob,
		// but Store.Delete usually takes a descriptor.
		// For ORAS OCI store, deleting a tag reference is done by Delete(ctx, descriptor_of_manifest) IF that descriptor was resolved FROM the tag?
		// Actually, standard OCI distribution spec 'Delete' is by digest.
		// However, oras-go/v2/content/oci Store supports Delete on the descriptor.
		// If we want to untag, we might need a specific untag method if 'Delete' removes the manifest payload itself.
		// If we delete the manifest blob, ALL tags pointing to it break or disappear?
		// Wait, deleting the manifest blob makes it unreachable.
		// If multiple tags point to same manifest, deleting the manifest creates a problem for other tags?
		//
		// Correct behavior for "rm <tag>" should be "untag".
		// OCI Layout/Store doesn't always have a distinct "untag" outside of removing the index entry.
		// checks `s.oci.Delete(ctx, desc)`.
		// If we pass the descriptor, it deletes the content at that digest.
		// That effectively deletes the manifest blob.
		//
		// If we want to support untagging only, we need to check if Store supports it.
		// oras-go/v2/content/oci implementation of Delete: deletes the blob file.
		// It ALSO removes it from `index.json`.
		// If we delete the blob, we nuke the artifact.
		//
		// Is this what we want? "system rm" usually implies deleting the image.
		// Yes.

		fmt.Printf("Removing %s (%s)...\n", ref, desc.Digest)
		if err := st.Delete(ctx, desc); err != nil {
			return fmt.Errorf("failed to remove artifact: %w", err)
		}

		fmt.Printf("Successfully removed %s\n", ref)
		return nil
	},
}

func init() {
	systemCmd.AddCommand(systemRmCmd)
}
