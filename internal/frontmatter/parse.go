package frontmatter

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNoFrontmatter is returned when a file does not start with a --- block.
var ErrNoFrontmatter = errors.New("no frontmatter found")

// File holds parsed frontmatter and the document body.
type File struct {
	Data map[string]interface{}
	body string
}

// Parse reads and parses the frontmatter of a markdown file.
func Parse(path string) (*File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(content)
}

// ParseBytes parses frontmatter from raw markdown bytes.
// Frontmatter must begin at byte 0 with "---\n" and close with "\n---\n" or "\n---" at EOF.
func ParseBytes(content []byte) (*File, error) {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return nil, ErrNoFrontmatter
	}
	rest := s[4:] // skip opening "---\n"

	var yamlContent, body string
	if idx := strings.Index(rest, "\n---\n"); idx >= 0 {
		yamlContent = rest[:idx]
		body = rest[idx+5:] // skip "\n---\n"
	} else if strings.HasPrefix(rest, "---\n") {
		// Handle empty frontmatter: ---\n---\n
		yamlContent = ""
		body = rest[4:] // skip "---\n"
	} else if strings.HasSuffix(rest, "\n---") {
		yamlContent = rest[:len(rest)-4]
		body = ""
	} else {
		return nil, ErrNoFrontmatter
	}

	data := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return nil, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	return &File{Data: data, body: body}, nil
}

// WriteFile serialises the (possibly modified) frontmatter back to disk.
// Note: yaml.v3 does not preserve original key order or inline comments.
func (f *File) WriteFile(path string) error {
	yamlBytes, err := yaml.Marshal(f.Data)
	if err != nil {
		return err
	}
	content := "---\n" + string(yamlBytes) + "---\n" + f.body
	return os.WriteFile(path, []byte(content), 0644)
}
