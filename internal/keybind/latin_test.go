package keybind

import "testing"

func TestLatinFoldsACyrillicKeyOntoItsPhysicalButton(t *testing.T) {
	for _, tc := range []struct{ typed, want string }{
		{"р", "h"},
		{"о", "j"},
		{"л", "k"},
		{"й", "q"},
		{"alt+р", "alt+h"},
		{"ctrl+я", "ctrl+z"},
		{"Р", "H"},
		{"б", ","},
		// Keys that are already latin, named or unmapped stay as they are.
		{"h", "h"},
		{"alt+h", "alt+h"},
		{"enter", "enter"},
		{"shift+up", "shift+up"},
		{"ctrl+\\", "ctrl+\\"},
		{"f3", "f3"},
		{" ", " "},
		{"", ""},
		{"привет", "привет"},
	} {
		if got := Latin(tc.typed); got != tc.want {
			t.Errorf("Latin(%q) = %q, want %q", tc.typed, got, tc.want)
		}
	}
}

// A binding answers to the physical key whatever the layout is on, and a
// capital typed in Russian folds the same way a shifted latin one does.
func TestNormalizeReadsARussianLayout(t *testing.T) {
	for _, tc := range []struct{ typed, want string }{
		{"о", "j"},
		{"alt+р", "alt+h"},
		{"Л", "K"},
		{"shift+о", "J"},
		{" ", "space"},
	} {
		if got := Normalize(tc.typed); got != tc.want {
			t.Errorf("Normalize(%q) = %q, want %q", tc.typed, got, tc.want)
		}
	}
}

// A key captured from the settings picker while the keyboard is Russian is
// stored as the latin spelling the config file and tmux both read.
func TestParseStoresACyrillicKeyAsItsLatinSpelling(t *testing.T) {
	key, err := Parse("alt+р")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if key.Tea() != "alt+h" || key.Tmux() != "M-h" {
		t.Fatalf("alt+р parsed as tea %q / tmux %q, want alt+h / M-h", key.Tea(), key.Tmux())
	}
	plain, err := Parse("о")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if plain.Tea() != "j" {
		t.Fatalf("о parsed as %q, want j", plain.Tea())
	}
}

// A tmux binding is registered by name, so an alt key needs the Russian
// character's name as well; ctrl and named keys arrive layout-independent
// and get none.
func TestTmuxTwinNamesTheRussianKeyForAltOnly(t *testing.T) {
	for _, tc := range []struct{ spec, want string }{
		{"alt+h", "M-р"},
		{"alt+q", "M-й"},
		{"alt+1", ""},
		{"ctrl+q", ""},
		{"f3", ""},
	} {
		key, err := Parse(tc.spec)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.spec, err)
		}
		if got := key.TmuxTwin(); got != tc.want {
			t.Errorf("Parse(%q).TmuxTwin() = %q, want %q", tc.spec, got, tc.want)
		}
	}
}
