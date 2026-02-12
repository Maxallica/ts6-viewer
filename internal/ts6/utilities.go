package ts6

import (
	"strings"
)

func UnescapeTS6(s string) string {
	replacer := strings.NewReplacer(
		`\s`, " ",
		`\p`, "|",
		`\/`, "/",
		`\:`, ":",
		`\.`, ".",
		`\(`, "(",
		`\)`, ")",
		`\?`, "?",
		`\!`, "!",
		`\-`, "-",
		`\_`, "_",
	)
	return replacer.Replace(s)
}
