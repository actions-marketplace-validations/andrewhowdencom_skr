package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoadMerged(t *testing.T) {
	// Setup Global Home
	globalDir := t.TempDir()
	t.Setenv("HOME", globalDir) // UserHomeDir uses HOME
	// Implementation uses os.UserConfigDir which uses XDG_CONFIG_HOME or HOME/.config
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(globalDir, ".config"))

	// Setup Project Hierarchy
	// /tmp/root/.skr.yaml
	// /tmp/root/subdir/project <--- LoadMerged called here
	rootDir := t.TempDir()
	middleDir := filepath.Join(rootDir, "subdir")
	projectDir := filepath.Join(middleDir, "project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// 1. Create Global Config (XDG)
	globalCfg := Config{
		Skills: []string{"global-skill"},
		Agents: []string{"antigravity"},
	}
	globalData, err := yaml.Marshal(globalCfg)
	require.NoError(t, err)

	globalConfigDir := filepath.Join(globalDir, ".config", "skr")
	err = os.MkdirAll(globalConfigDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(globalConfigDir, "config.yaml"), globalData, 0644)
	require.NoError(t, err)

	// 2. Create Local Config in ROOT (Parent of project)
	localCfg := Config{
		Skills: []string{"local-skill"},
		Agents: []string{"roocode"},
	}
	localData, err := yaml.Marshal(localCfg)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(rootDir, ".skr.yaml"), localData, 0644)
	require.NoError(t, err)

	// 3. Test LoadMerged from deep project dir
	cfg, err := LoadMerged(projectDir)
	require.NoError(t, err)

	// Verify Skills (Appended)
	assert.Contains(t, cfg.Skills, "global-skill")
	assert.Contains(t, cfg.Skills, "local-skill")

	// Verify Agents (Merged)
	assert.Contains(t, cfg.Agents, "antigravity")
	assert.Contains(t, cfg.Agents, "roocode")
	assert.Equal(t, 2, len(cfg.Agents))
}
