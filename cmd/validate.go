package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate an Agent Skill definition",
	Long: `Validate an Agent Skill's integrity and adherence to the specification.

Checks for:
- Existence of SKILL.md
- Valid frontmatter
- Spec compliance (naming, fields)
- Directory structure

If [path] is not provided, defaults to the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		s, err := skill.Load(path)
		if err != nil {
			return fmt.Errorf("skill is invalid: %w", err)
		}

		fmt.Printf("Skill '%s' is valid.\n", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
