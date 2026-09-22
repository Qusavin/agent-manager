package keybind

import "strings"

// latinByCyrillic maps every character the Russian ЙЦУКЕН layout types to
// the one the same physical key types on a US layout. A keyboard switched
// to Russian reports "р" where a binding is written "h", so the table is
// read the other way round to recover the key that was actually pressed.
var latinByCyrillic = map[rune]rune{
	'ё': '`', 'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y',
	'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
	'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j',
	'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
	'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n', 'ь': 'm',
	'б': ',', 'ю': '.',

	'Ё': '~', 'Й': 'Q', 'Ц': 'W', 'У': 'E', 'К': 'R', 'Е': 'T', 'Н': 'Y',
	'Г': 'U', 'Ш': 'I', 'Щ': 'O', 'З': 'P', 'Х': '{', 'Ъ': '}',
	'Ф': 'A', 'Ы': 'S', 'В': 'D', 'А': 'F', 'П': 'G', 'Р': 'H', 'О': 'J',
	'Л': 'K', 'Д': 'L', 'Ж': ':', 'Э': '"',
	'Я': 'Z', 'Ч': 'X', 'С': 'C', 'М': 'V', 'И': 'B', 'Т': 'N', 'Ь': 'M',
	'Б': '<', 'Ю': '>',
}

// Latin rewrites a key the terminal reported in Cyrillic as the key its
// physical button carries on a US layout, so bindings keep answering while
// the keyboard is switched to Russian. Modifiers are kept and only the
// character behind them is folded; anything else - a named key, a key
// already latin - comes back untouched.
//
// Ctrl+<letter> needs none of this: the terminal sends those as control
// bytes chosen by the physical key, layout or no layout. Alt+<letter> and
// plain characters arrive as text, which is where a switched layout shows.
func Latin(key string) string {
	mods, rest := "", key
	if plus := strings.LastIndex(key, "+"); plus > 0 && plus+1 < len(key) {
		mods, rest = key[:plus+1], key[plus+1:]
	}
	runes := []rune(rest)
	if len(runes) != 1 {
		return key
	}
	latin, mapped := latinByCyrillic[runes[0]]
	if !mapped {
		return key
	}
	return mods + string(latin)
}

// cyrillicByLatin is latinByCyrillic read backwards. Surfaces that register
// a key by name rather than matching one - a tmux binding is the case - have
// to name the Russian character too: tmux compares what the terminal sends,
// and a switched layout sends the other character.
var cyrillicByLatin = func() map[rune]rune {
	back := make(map[rune]rune, len(latinByCyrillic))
	for cyrillic, latin := range latinByCyrillic {
		back[latin] = cyrillic
	}
	return back
}()

// TmuxTwin is the tmux name this key answers to while the keyboard is on a
// Russian layout, empty when it needs none. Only alt keys do: tmux reads
// them as the character the layout types, where ctrl keys and named keys
// arrive as themselves.
func (k Key) TmuxTwin() string {
	rest, isAlt := strings.CutPrefix(k.tmux, "M-")
	if !isAlt {
		return ""
	}
	runes := []rune(rest)
	if len(runes) != 1 {
		return ""
	}
	cyrillic, mapped := cyrillicByLatin[runes[0]]
	if !mapped {
		return ""
	}
	return "M-" + string(cyrillic)
}
