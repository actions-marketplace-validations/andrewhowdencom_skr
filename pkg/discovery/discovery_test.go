package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListInstalledSkills(t *testing.T) {
	// Setup Dirs
	projectRoot := t.TempDir()

	agentDir := filepath.Join(projectRoot, ".agent", "skills")
	err := os.MkdirAll(agentDir, 0755)
	require.NoError(t, err)

	extraDir := t.TempDir()

	globalHome := t.TempDir()
	t.Setenv("HOME", globalHome)
	globalSkillsDir := filepath.Join(globalHome, ".config", "agent", "skills")
	err = os.MkdirAll(globalSkillsDir, 0755)
	require.NoError(t, err)

	// Helper to create skill
	createSkill := func(dir, name string) {
		skillDir := filepath.Join(dir, name)
		err := os.MkdirAll(skillDir, 0755)
		require.NoError(t, err)
		// Minimal SKILL.md with valid frontmatter
		content := "---\nname: " + name + "\ndescription: test skill\n---\n"
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
		require.NoError(t, err)
	}

	// 1. Create Skills
	// "local-only" in .agent/skills
	createSkill(agentDir, "local-only")

	// "extra-only" in extraDir
	createSkill(extraDir, "extra-only")

	// "global-only" in globalDir
	createSkill(globalSkillsDir, "global-only")

	// "override-extra": in local AND extra (local should win, IsGlobal=false)
	createSkill(agentDir, "override-extra")
	createSkill(extraDir, "override-extra")

	// "override-global": in extra AND global (extra should win, IsGlobal=true, but coming from extra path)
	createSkill(extraDir, "override-global")
	createSkill(globalSkillsDir, "override-global")

	// "override-all": in local, extra, global (local should win)
	createSkill(agentDir, "override-all")
	createSkill(extraDir, "override-all")
	createSkill(globalSkillsDir, "override-all")

	// 2. Run ListInstalledSkills
	skills, err := ListInstalledSkills(projectRoot, []string{extraDir})
	require.NoError(t, err)

	// 3. Verify
	stats := make(map[string]InstalledSkill)
	for _, s := range skills {
		stats[s.Name] = s
	}

	// Check counts
	// local-only, extra-only, global-only, override-extra, override-global, override-all = 6 unique skills
	assert.Equal(t, 6, len(skills))

	// local-only
	s, ok := stats["local-only"]
	require.True(t, ok)
	assert.False(t, s.IsGlobal)
	assert.Contains(t, s.Path, agentDir)

	// extra-only
	s, ok = stats["extra-only"]
	require.True(t, ok)
	assert.True(t, s.IsGlobal) // Treated as global/external
	assert.Contains(t, s.Path, extraDir)

	// global-only
	s, ok = stats["global-only"]
	require.True(t, ok)
	assert.True(t, s.IsGlobal)
	assert.Contains(t, s.Path, globalSkillsDir)

	// override-extra (Local wins)
	s, ok = stats["override-extra"]
	require.True(t, ok)
	assert.False(t, s.IsGlobal)
	assert.Contains(t, s.Path, agentDir)

	// override-global (Extra wins)
	s, ok = stats["override-global"]
	require.True(t, ok)
	assert.True(t, s.IsGlobal)
	assert.Contains(t, s.Path, extraDir)

	// override-all (Local wins)
	s, ok = stats["override-all"]
	require.True(t, ok)
	assert.False(t, s.IsGlobal)
	assert.Contains(t, s.Path, agentDir)
}
