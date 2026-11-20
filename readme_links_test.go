package linter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	readmeLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)[^)]*\)`)
	headingPattern    = regexp.MustCompile(`^#{1,6}\s+(.*)$`)
	inlineCodePattern = regexp.MustCompile("`[^`]*`")
)

func githubAnchor(heading string) string {
	var anchor strings.Builder

	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			anchor.WriteRune('-')
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			anchor.WriteRune(r)
		}
	}

	return anchor.String()
}

func headingAnchorsOf(content string) map[string]bool {
	anchors := make(map[string]bool)

	inFence := false

	for line := range strings.SplitSeq(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence

			continue
		}

		if inFence {
			continue
		}

		if m := headingPattern.FindStringSubmatch(line); m != nil {
			anchors["#"+githubAnchor(strings.TrimSpace(m[1]))] = true
		}
	}

	return anchors
}

func readAnchors(root, rel string) (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rel, err)
	}

	return headingAnchorsOf(string(data)), nil
}

func missingAnchor(anchors map[string]bool, target, frag string) []string {
	if anchors["#"+frag] {
		return nil
	}

	return []string{fmt.Sprintf("no README heading produces anchor %q", target)}
}

func fileLinkFailures(root string, target, path, frag string, hasFrag bool) []string {
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
		return []string{fmt.Sprintf("target %q missing on disk", path)}
	}

	if !hasFrag || !strings.HasSuffix(path, ".md") {
		return nil
	}

	subAnchors, err := readAnchors(root, path)
	if err != nil {
		return []string{err.Error()}
	}

	if subAnchors["#"+frag] {
		return nil
	}

	return []string{fmt.Sprintf("no heading in %s produces anchor %q", path, target)}
}

func linkFailures(root string, anchors map[string]bool, target string) []string {
	path, frag, hasFrag := strings.Cut(target, "#")

	switch {
	case path == "" && hasFrag:
		return missingAnchor(anchors, target, frag)
	case strings.Contains(path, "://"), strings.HasPrefix(path, "mailto:"):
		return nil
	default:
		return fileLinkFailures(root, target, path, frag, hasFrag)
	}
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

	anchors := headingAnchorsOf(string(data))

	var broken []string

	inFence := false

	for i, rawLine := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(rawLine), "```") {
			inFence = !inFence

			continue
		}

		if inFence {
			continue
		}

		code := inlineCodePattern.ReplaceAllString(rawLine, "")

		for _, m := range readmeLinkPattern.FindAllStringSubmatch(code, -1) {
			for _, failure := range linkFailures(root, anchors, m[1]) {
				broken = append(broken, fmt.Sprintf("README.md:%d %q: %s", i+1, m[1], failure))
			}
		}
	}

	if len(broken) > 0 {
		t.Errorf("README.md has %d broken relative link(s):\n%s", len(broken), strings.Join(broken, "\n"))
	}
}
