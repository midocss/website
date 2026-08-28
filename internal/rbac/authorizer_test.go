package rbac

import (
	"slices"
	"testing"

	"github.com/midocss/website/internal/domain"
)

func effect(value string) *string { return &value }

func TestEffectiveDropsDeniedPermissions(t *testing.T) {
	rows := []permissionRow{
		{Slug: "packages.view"},
		{Slug: "packages.create", Effect: effect(domain.PermissionEffectAllow)},
		{Slug: "packages.delete", Effect: effect(domain.PermissionEffectDeny)},
	}

	got := effective(rows)
	want := []string{"packages.view", "packages.create"}
	if !slices.Equal(got, want) {
		t.Fatalf("effective() = %v, want %v", got, want)
	}
}

func TestHasAll(t *testing.T) {
	granted := []string{"users.view", "users.create"}

	tests := []struct {
		name     string
		required []string
		want     bool
	}{
		{name: "single granted", required: []string{"users.view"}, want: true},
		{name: "all granted", required: []string{"users.view", "users.create"}, want: true},
		{name: "one missing", required: []string{"users.view", "users.delete"}, want: false},
		{name: "none required", required: nil, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasAll(granted, tc.required); got != tc.want {
				t.Fatalf("hasAll(%v) = %v, want %v", tc.required, got, tc.want)
			}
		})
	}
}
