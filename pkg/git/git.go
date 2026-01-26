package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetShortSHA returns the short SHA of the current HEAD
func GetShortSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GetHeadTags returns a list of git tags pointing to the current HEAD
func GetHeadTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "--points-at", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var tags []string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			tags = append(tags, strings.TrimSpace(line))
		}
	}
	return tags, nil
}

// ChangedFiles returns a list of files changed between HEAD and baseRef
func ChangedFiles(baseRef string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", baseRef)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff failed: %s (%w)", out.String(), err)
	}

	lines := strings.Split(out.String(), "\n")
	var files []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			files = append(files, strings.TrimSpace(line))
		}
	}
	return files, nil
}
