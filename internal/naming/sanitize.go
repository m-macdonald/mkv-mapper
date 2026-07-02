package naming

import "strings"

var charReplacements = map[rune]string{
	'<': "",
	'>': "",
	':': " - ",
	'"': "'",
	'/': "-",
	'\\': "-",
	'|': "-",
	'?': "",
	'*': "",
}

func sanitizeSegment(name string) string {
	var b strings.Builder
	for _, r := range name {
		if replacement, ok := charReplacements[r]; ok {
			b.WriteString(replacement)
			continue
		}
		b.WriteRune(r)
	}


	s := b.String()
	s = strings.Join(strings.Fields(s), " ")

	if s == "" {
		s = "_"
	}
	return s
}
