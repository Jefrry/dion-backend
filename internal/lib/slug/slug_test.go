package slug

import "testing"

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Привет, мир!":                            "privet-mir",
		"Тестовый заголовок статьи":               "testovyy-zagolovok-stati",
		"Go + русский язык = круто":               "go-russkiy-yazyk-kruto",
		"  Это   строка   с   пробелами  ":        "eto-stroka-s-probelami",
		"Съешь ещё этих мягких французских булок": "sesh-eshe-etih-myagkih-frantsuzskih-bulok",
		"Hello World!":        "hello-world",
		"Mixed Текст 123":     "mixed-tekst-123",
		"---Already--slug---": "already-slug",
	}

	for input, expected := range tests {
		result := Slugify(input)
		if result != expected {
			t.Errorf("Slugify(%q) = %q; want %q", input, result, expected)
		}
	}
}
