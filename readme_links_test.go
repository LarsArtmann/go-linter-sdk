package linter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var readmeLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)[^)]*\)`)

var headingPattern = regexp.MustCompile(`^#{1,6}\s+(.*)$`)

func githubAnchor(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}

func unfencedLines(content string) []string {
	var kept []string
	inFence := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if !inFence {
			kept = append(kept, line)
		}
	}
	return kept
}

func anchorsOf(lines []string) map[string]bool {
	anchors := make(map[string]bool)
	for _, line := range lines {
		if m := headingPattern.FindStringSubmatch(line); m != nil {
			anchors["#"+githubAnchor(strings.TrimSpace(m[1]))] = true
		}
	}
	return anchors
}

func fileAnchors(root, rel string) (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	return anchorsOf(unfencedLines(string(data))), nil
}

func TestReadmeRelativeLinksResolve(t *testing.T) {
	t.Parallel()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	lines := unfencedLines(string(data))
	anchors := anchorsOf(lines)

	var broken []string
	for i, line := range lines {
		for _, m := range readmeLinkPattern.FindAllStringSubmatch(line, -1) {
			target := m[1]
			path, frag, hasFrag := strings.Cut(target, "#")
			where := fmt.Sprintf("README.md:%d %q", i+1, target)

			switch {
			case path == "" && hasFrag:
				if !anchors["#"+frag] {
					broken = append(broken, where+": no README heading produces this anchor")
				}
			case strings.Contains(path, "://"), strings.HasPrefix(path, "mailto:"):
			default:
				if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
					broken = append(broken, where+": target missing on disk")
					continue
				}
				if hasFrag && strings.HasSuffix(path, ".md") {
					subAnchors, err := fileAnchors(root, path)
					if err != nil {
						broken = append(broken, where+": unreadable: "+err.Error())
						continue
					}
					if !subAnchors["#"+frag] {
						broken = append(broken, where+": no heading in "+path+" produces this anchor")
					}
				}
			}
		}
	}

	if len(broken) > 0 {
		t.Errorf("README.md has %d broken relative link(s):\n%s", len(broken), strings.Join(broken, "\n"))
	}
}
