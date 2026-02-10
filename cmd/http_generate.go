package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/andrewhowdencom/skr/pkg/ui"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

var outputDir string

var httpGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the static UI site",
	Long:  `Generate a static HTML website and skills.json representing the keys in the OCI store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// 1. Initialize Store
		st, err := store.New(ociPath)
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 2. List Skills from Store
		fmt.Println("Scanning OCI store for skills...")
		tags, err := st.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list skills from store: %w", err)
		}

		// REFACTOR: Use a custom struct for UI JSON to include description/author
		type UISkill struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Path        string `json:"path,omitempty"` // Optional for OCI
			Metadata    struct {
				Author  string `json:"author"`
				Version string `json:"version"`
			} `json:"metadata"`
		}

		var uiSkills []UISkill
		for _, tag := range tags {
			desc, err := st.Resolve(ctx, tag)
			if err != nil {
				continue
			}

			rc, err := st.Fetch(ctx, desc)
			if err != nil {
				continue
			}
			containerBytes, _ := io.ReadAll(rc)
			rc.Close()

			var manifest v1.Manifest
			json.Unmarshal(containerBytes, &manifest)

			uiSkills = append(uiSkills, UISkill{
				Name:        tag, // or parse name from tag
				Description: manifest.Annotations["com.skr.description"],
				Metadata: struct {
					Author  string `json:"author"`
					Version string `json:"version"`
				}{
					Author:  manifest.Annotations["com.skr.author"],
					Version: manifest.Annotations["com.skr.version"],
				},
			})
		}

		fmt.Printf("Found %d skills in store.\n", len(uiSkills))

		// 3. Create Output Directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// 4. Write skills.json
		skillsPath := filepath.Join(outputDir, "skills.json")
		file, err := os.Create(skillsPath)
		if err != nil {
			return fmt.Errorf("failed to create skills.json: %w", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(uiSkills); err != nil {
			return fmt.Errorf("failed to encode skills.json: %w", err)
		}
		fmt.Printf("Written %s\n", skillsPath)

		// 5. Extract Embedded Assets
		assets, err := ui.Assets()
		if err != nil {
			return fmt.Errorf("failed to load embedded assets: %w", err)
		}

		filesToCopy := []string{"index.html", "style.css", "app.js"}

		for _, fileName := range filesToCopy {
			f, err := assets.Open(fileName)
			if err != nil {
				return fmt.Errorf("failed to open embedded asset %s: %w", fileName, err)
			}

			destPath := filepath.Join(outputDir, fileName)
			destFile, err := os.Create(destPath)
			if err != nil {
				f.Close()
				return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
			}

			_, err = io.Copy(destFile, f)
			f.Close()
			destFile.Close()
			if err != nil {
				return fmt.Errorf("failed to copy asset %s: %w", fileName, err)
			}
			fmt.Printf("Written %s\n", destPath)
		}

		fmt.Println("HTTP generation complete.")
		return nil
	},
}

func init() {
	httpCmd.AddCommand(httpGenerateCmd)
	httpGenerateCmd.Flags().StringVarP(&outputDir, "output", "o", "build/http", "Directory to output the generated site")
}
