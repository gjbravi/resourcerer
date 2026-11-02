package auth

import "slices"

// Access resolves what a member of groups may do. allowedGroups is the perimeter: outside it a user has no access,
// whatever the roles say. An empty editorGroups means nobody may change anything.
func (c Config) Access(groups []string) (viewer, editor bool) {
	if len(c.AllowedGroups) > 0 && !intersects(groups, c.AllowedGroups) {
		return false, false
	}
	editor = intersects(groups, c.Roles.EditorGroups)
	return editor || c.Roles.anyViewer(), editor
}

// referenced keeps the groups the configuration mentions, to fit a session cookie for a user in many groups.
func (c Config) referenced(groups []string) []string {
	out := []string{}
	for _, g := range groups {
		if slices.Contains(c.AllowedGroups, g) || slices.Contains(c.Roles.EditorGroups, g) {
			out = append(out, g)
		}
	}
	return out
}

func (r Roles) anyViewer() bool {
	return r.AnyAuthenticatedIsViewer == nil || *r.AnyAuthenticatedIsViewer
}

func intersects(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}
