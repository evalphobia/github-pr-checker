package prchecker

import "strings"

// Comment contains comments to add to pull request.
type Comment struct {
	list []string
}

// HasComment checks if at least one comment is exists or not.
func (c *Comment) HasComment() bool {
	return len(c.list) != 0
}

// Add adds comment into the list.
func (c *Comment) Add(s string) {
	if s == "" {
		return
	}

	c.list = append(c.list, s)
}

// Show gets joined comment.
func (c *Comment) Show() string {
	return strings.Join(c.list, "\n\n")
}
