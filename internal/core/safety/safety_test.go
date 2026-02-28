package safety_test

import (
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/safety"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// SG-001
		{name: "valid uuid v4 lowercase", input: "86334621-8152-4e67-b322-9f139d6c0a57", wantErr: false},
		// SG-002
		{name: "valid uuid v4 uppercase treated as valid", input: "86334621-8152-4E67-B322-9F139D6C0A57", wantErr: false},
		// SG-003
		{name: "random string", input: "not-a-uuid", wantErr: true},
		// SG-004
		{name: "uuid v1 rejected", input: "550e8400-e29b-11d4-a716-446655440000", wantErr: true},
		// SG-005
		{name: "empty string", input: "", wantErr: true},
		{name: "too short", input: "86334621-8152-4e67-b322", wantErr: true},
		{name: "version 5 rejected", input: "886313e1-3b8a-5372-9b90-0c9aee199e5d", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := safety.ValidateUUID(tc.input)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for input %q, got nil", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for input %q: %v", tc.input, err)
			}
		})
	}
}

func TestValidateDescendant(t *testing.T) {
	root := "/home/user/.copilot/session-state"

	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		// SG-006
		{name: "valid descendant", target: root + "/86334621-8152-4e67-b322-9f139d6c0a57", wantErr: false},
		// SG-007
		{name: "path traversal", target: root + "/../../etc/passwd", wantErr: true},
		// SG-008
		{name: "target is root itself", target: root, wantErr: true},
		// SG-011
		{name: "empty target", target: "", wantErr: true},
		{name: "empty root", target: root + "/uuid", wantErr: true}, // root="" handled
		{name: "sibling directory", target: "/home/user/.copilot/other", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := root
			if tc.name == "empty root" {
				r = ""
			}
			err := safety.ValidateDescendant(r, tc.target)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for target %q, got nil", tc.target)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for target %q: %v", tc.target, err)
			}
		})
	}
}
