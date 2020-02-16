package files

import (
	"path/filepath"

	ignore "github.com/crackcomm/go-gitignore"
)

// Filter - Filter filter.
type Filter struct {
	// IncludeHidden - Include hidden files
	IncludeHidden bool
	// Rules - File filter rules
	Rules *ignore.GitIgnore
}

// NewFilter creates a new file filter form .gitignore and list of rules.
func NewFilter(ignoreFile string, rules []string, includeHidden bool) (filter *Filter, err error) {
	filter = &Filter{IncludeHidden: includeHidden}
	if ignoreFile == "" {
		filter.Rules, err = ignore.CompileIgnoreLines(rules...)
	} else {
		filter.Rules, err = ignore.CompileIgnoreFileAndLines(ignoreFile, rules...)
	}
	return
}

func isPathHidden(path string) bool {
	path = filepath.Base(path)
	if path == "." || path == "" {
		return false
	}
	if rune(path[0]) == rune('.') {
		return true
	}
	return false
}

// Filter returns true if a file should be filtered
func (filter *Filter) Filter(path string) bool {
	if isPathHidden(path) {
		if filter.IncludeHidden {
			return false
		}
		return true
	}
	return filter.Rules.MatchesPath(path)
}
