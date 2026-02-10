package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/andrewhowdencom/skr/pkg/ui"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

var port int

var httpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Skills Registry UI locally",
	Long:  `Start a local HTTP server that builds and serves the UI on-the-fly, reflecting the current state of the OCI store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// 1. Initialize Store
		st, err := store.New(ociPath)
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 2. Asset Handler
		assets, err := ui.Assets()
		if err != nil {
			return fmt.Errorf("failed to load embedded assets: %w", err)
		}

		fileServer := http.FileServer(http.FS(assets))

		// 3. API Handler
		http.HandleFunc("/api/skills", func(w http.ResponseWriter, r *http.Request) {
			// Scan OCI store on every request
			tags, err := st.List(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list skills: %v", err), http.StatusInternalServerError)
				return
			}

			type SkillVersion struct {
				Version string `json:"version"`
				Tag     string `json:"tag"`
			}

			type UISkill struct {
				ID          string         `json:"id"`   // Full repo name (e.g. ghcr.io/user/skill)
				Name        string         `json:"name"` // Short name (e.g. skill)
				Description string         `json:"description"`
				Author      string         `json:"author"`
				Versions    []SkillVersion `json:"versions"`
				LatestTag   string         `json:"latestTag"`
			}

			skillMap := make(map[string]*UISkill)

			for _, tag := range tags {
				// Parse Tag: repo:version
				// If no colon, assume latest? (OCI spec usually implies a tag)
				// Simple parsing: last colon is the separator
				repo := tag
				version := "latest"
				if lastIdx := lastIndex(tag, ":"); lastIdx != -1 {
					repo = tag[:lastIdx]
					version = tag[lastIdx+1:]
				}

				// Resolve to get metadata
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

				// Get or Create Skill Entry
				if _, exists := skillMap[repo]; !exists {
					// Derive simplified name
					shortName := repo
					if lastSlash := lastIndex(repo, "/"); lastSlash != -1 {
						shortName = repo[lastSlash+1:]
					}

					skillMap[repo] = &UISkill{
						ID:          repo,
						Name:        shortName,
						Description: manifest.Annotations["com.skr.description"], // First one wins for desc
						Author:      manifest.Annotations["com.skr.author"],      // First one wins for author
						Versions:    []SkillVersion{},
					}
				}

				// Append Version
				// If annotation version is present, use it, else use tag (which we parsed above)
				displayVersion := version
				if v := manifest.Annotations["com.skr.version"]; v != "" {
					displayVersion = v
				}

				skillMap[repo].Versions = append(skillMap[repo].Versions, SkillVersion{
					Version: displayVersion,
					Tag:     tag,
				})

				// Update metadata if missing (e.g. if we processed a tag with empty metadata first)
				if skillMap[repo].Description == "" {
					skillMap[repo].Description = manifest.Annotations["com.skr.description"]
				}
				if skillMap[repo].Author == "" {
					skillMap[repo].Author = manifest.Annotations["com.skr.author"]
				}
			}

			// Convert Map to Slice
			var uiSkills []UISkill
			for _, s := range skillMap {
				uiSkills = append(uiSkills, *s)
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(uiSkills); err != nil {
				http.Error(w, "Failed to encode skills", http.StatusInternalServerError)
			}
		})

		http.HandleFunc("/skills.json", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/api/skills", http.StatusTemporaryRedirect)
		})

		http.Handle("/", fileServer)

		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Serving Skills Registry (HTTP) at http://localhost%s\n", addr)
		if ociPath == "" {
			fmt.Println("Source: System OCI Store")
		} else {
			fmt.Printf("Source: %s\n", ociPath)
		}
		fmt.Println("Press Ctrl+C to stop")

		if err := http.ListenAndServe(addr, nil); err != nil {
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	},
}

func init() {
	httpCmd.AddCommand(httpServeCmd)
	httpServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
}
