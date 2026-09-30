package snapshot

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/neox5/snp/internal/config"
)

// starterGitignore is the .gitignore of the ssh-public-keys-starter repository.
const starterGitignore = `# Block everything
*

# Allow only specific directories
!automation/
!personal/
!emergency/

# Allow only specific files
!automation/*.pub
!personal/*.pub
!emergency/*.pub

# Allow placeholders that keep empty directories tracked
!.gitkeep

# Allow documentation
!README.md
!.gitignore
`

// TestCollect_StarterRepo is the original bug: the allowlist .gitignore of the
// starter repository produced an empty snapshot.
func TestCollect_StarterRepo(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		".gitignore":                        starterGitignore,
		"README.md":                         "readme",
		"automation/.gitkeep":               "",
		"automation/id_ed25519_ansible":     "PRIVATE",
		"automation/id_ed25519_ansible.pub": "ssh-ed25519 AAA ansible",
		"personal/.gitkeep":                 "",
		"personal/id_ed25519_laptop":        "PRIVATE",
		"emergency/.gitkeep":                "",
		"sub/.gitkeep":                      "",
		"sub/x":                             "x",
	}
	for name, content := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := collect(&config.Config{SourceDir: root, Depth: -1}, root)
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, e := range entries {
		got = append(got, e.RelPath)
	}
	sort.Strings(got)

	// same as: git ls-files --others --exclude-standard
	want := []string{
		".gitignore",
		"README.md",
		"automation/.gitkeep",
		"automation/id_ed25519_ansible.pub",
		"emergency/.gitkeep",
		"personal/.gitkeep",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collected %v, want %v", got, want)
	}
}
