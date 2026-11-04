package auth

import "testing"

func TestAccess(t *testing.T) {
	no := false
	for _, tc := range []struct {
		name           string
		cfg            Config
		groups         []string
		viewer, editor bool
	}{
		{"no groups configured: viewer", Config{}, []string{"dev"}, true, false},
		{"no groups at all: viewer", Config{}, nil, true, false},
		{"editor group", Config{Roles: Roles{EditorGroups: []string{"sre"}}}, []string{"dev", "sre"}, true, true},
		{"empty editorGroups: nobody edits", Config{Roles: Roles{EditorGroups: nil}}, []string{"sre"}, true, false},
		{"outside allowedGroups", Config{AllowedGroups: []string{"platform"}}, []string{"dev"}, false, false},
		{"allowedGroups beats editorGroups", Config{AllowedGroups: []string{"platform"},
			Roles: Roles{EditorGroups: []string{"sre"}}}, []string{"sre"}, false, false},
		{"inside allowedGroups", Config{AllowedGroups: []string{"platform"}}, []string{"platform"}, true, false},
		{"viewers off: non-editor has no access", Config{Roles: Roles{AnyAuthenticatedIsViewer: &no}}, []string{"dev"}, false, false},
		{"viewers off: editor keeps access", Config{Roles: Roles{AnyAuthenticatedIsViewer: &no,
			EditorGroups: []string{"sre"}}}, []string{"sre"}, true, true},
	} {
		viewer, editor := tc.cfg.Access(tc.groups)
		if viewer != tc.viewer || editor != tc.editor {
			t.Errorf("%s: viewer %v editor %v, want %v %v", tc.name, viewer, editor, tc.viewer, tc.editor)
		}
	}
}

func TestReferencedGroups(t *testing.T) {
	c := Config{AllowedGroups: []string{"platform"}, Roles: Roles{EditorGroups: []string{"sre"}}}
	got := c.referenced([]string{"a", "sre", "b", "platform"})
	if len(got) != 2 || got[0] != "sre" || got[1] != "platform" {
		t.Errorf("got %v", got)
	}
}
