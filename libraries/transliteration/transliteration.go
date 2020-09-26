package transliteration

import (
	"bufio"
	"io"
	"log"

	"github.com/alexsergivan/transliterator"
)

func NewTransliterator(overwrites map[rune]string) *Transliterator {
	customLanguageOverrites := make(map[string]map[rune]string)
	customLangcode := "custom_langcode"
	customLanguageOverrites[customLangcode] = overwrites
	trans := transliterator.NewTransliterator(&customLanguageOverrites)
	return &Transliterator{
		tl:       trans,
		langcode: customLangcode,
	}
}

func NewDanishTransliterator() *Transliterator {
	danishOverwrites := make(map[string]map[rune]string)
	danishLangcode := "daCustom"
	danishOverwrites[danishLangcode] = map[rune]string{
		0x00E6: "æ",
		0x00F8: "ø",
		0x00E5: "å",
		0x00C6: "Æ",
		0x00D8: "Ø",
		0x00C5: "Å",
		0x00A7: "§",
	}
	trans := transliterator.NewTransliterator(&danishOverwrites)
	return &Transliterator{
		tl:       trans,
		langcode: danishLangcode,
	}
}

type Transliterator struct {
	tl       *transliterator.Transliterator
	langcode string
}

func (t *Transliterator) Process(r *bufio.Reader, w *bufio.Writer) {
	for {
		// TODO: line break might be carriage return or other chars
		switch line, err := r.ReadString('\n'); err {

		case io.EOF:
			cleanedLine := t.tl.Transliterate(line, t.langcode)
			_, err := w.WriteString(cleanedLine)
			if err != nil {
				log.Fatal(err)
			}
			w.Flush()
			return

		case nil:
			cleanedLine := t.tl.Transliterate(line, t.langcode)
			_, err := w.WriteString(cleanedLine)
			if err != nil {
				log.Fatal(err)
			}

		default:
			w.Flush()
			log.Fatal(err)
			return
		}
	}

}
