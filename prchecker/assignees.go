package prchecker

import (
	"sync"
)

// Assignees has GitHub users.
type Assignees struct {
	once sync.Once
	// listMap contains GitHub users in Key.
	listMap map[string]struct{}
}

// HasAssignees checks if at least one user is exists or not.
func (a *Assignees) HasAssignees() bool {
	return len(a.listMap) != 0
}

// RemoveFromList removes users from the list.
func (a *Assignees) RemoveFromList(names ...string) {
	for _, name := range names {
		if _, ok := a.listMap[name]; ok {
			delete(a.listMap, name)
		}
	}
}

// Add adds users into the list.
func (a *Assignees) Add(assignees ...string) {
	if len(assignees) == 0 {
		return
	}

	a.once.Do(func() {
		a.listMap = make(map[string]struct{})
	})

	for _, assignee := range assignees {
		a.listMap[assignee] = struct{}{}
	}
}

// List returns the list of users.
func (a *Assignees) List() []string {
	size := len(a.listMap)
	if size == 0 {
		return nil
	}

	list := make([]string, 0, size)
	for key := range a.listMap {
		list = append(list, key)
	}
	return list
}
