package domain

import (
	"strings"
	"testing"
)

func TestProjectValidate(t *testing.T) {
	cases := []struct {
		name    string
		pName   string
		pSlug   string
		wantErr bool
	}{
		{"valid", "My Project", "my-project", false},
		{"empty name", "", "my-project", true},
		{"spaces name", "   ", "my-project", true},
		{"long name", strings.Repeat("a", 201), "my-project", true},
		{"empty slug", "My Project", "", true},
		{"invalid slug uppercase", "My Project", "My-Project", true},
		{"invalid slug symbols", "My Project", "my_project", true},
		{"valid slug numbers", "Project 123", "proj-123", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := Project{Name: tc.pName, Slug: tc.pSlug}
			err := p.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
