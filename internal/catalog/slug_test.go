package catalog

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "E-Commerce Store", want: "e-commerce-store"},
		{in: "  Landing   Page!! ", want: "landing-page"},
		{in: "متجر إلكتروني", want: "متجر-إلكتروني"},
		{in: "---", want: ""},
		{in: "Plan #2 (Pro)", want: "plan-2-pro"},
	}

	for _, tc := range tests {
		if got := Slugify(tc.in); got != tc.want {
			t.Errorf("Slugify(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
