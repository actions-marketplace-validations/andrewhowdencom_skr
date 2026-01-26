package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/andrewhowdencom/skr/pkg/discovery"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Agent Skills",
	Long: `List all Agent Skills currently installed in the active agent context (project).

Scans the hierarchy for .agent/skills and merges with global skills.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		skills, err := discovery.ListInstalledSkills(cwd)
		if err != nil {
			// If err means not found, behave gracefully
			fmt.Printf("No agent context found (searching up from %s).\n", cwd)
			return nil
		}

		if len(skills) == 0 {
			fmt.Println("No skills installed in this context.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tGLOBAL\tPATH")

		for _, s := range skills {
			globalMark := ""
			if s.IsGlobal {
				globalMark = "*"
			}

			// Relative path if possible for cleaner output
			displayPath := s.Path
			if rel, err := filepath.Rel(cwd, s.Path); err == nil {
				// Use relative path only if it's shorter or reasonable?
				// Actually rel path is usually preferred for local context.
				displayPath = rel
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.Version, globalMark, displayPath)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
