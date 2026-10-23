package operatingsystem

import "testing"

func TestUnquote(t *testing.T) {
	for in, want := range map[string]string{
		`"BurmillaOS v2.0.1"`: "BurmillaOS v2.0.1",
		`'Debian GNU/Linux'`:  "Debian GNU/Linux",
		`Alpine`:              "Alpine",
		`"with \"esc\""`:      `with "esc"`,
	} {
		if got := unquote(in); got != want {
			t.Errorf("unquote(%q) = %q, want %q", in, got, want)
		}
	}
}
