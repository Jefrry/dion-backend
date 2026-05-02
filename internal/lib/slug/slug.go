package slug

import (
	"regexp"
	"strings"
)

var (
	nonAlnumRegexp  = regexp.MustCompile(`[^a-z0-9]+`)
	multiDashRegexp = regexp.MustCompile(`-+`)
)

// Таблица транслитерации
var ruToEn = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
	'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch",
	'ш': "sh", 'щ': "sh", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya",
}

// Slugify converts string to URL-friendly slug
func Slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))

	var b strings.Builder

	for _, r := range input {
		// Русские символы
		if val, ok := ruToEn[r]; ok {
			b.WriteString(val)
			continue
		}

		// Латиница и цифры
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}

		// Остальное → дефис
		b.WriteRune('-')
	}

	slug := b.String()

	slug = nonAlnumRegexp.ReplaceAllString(slug, "-")
	slug = multiDashRegexp.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}
