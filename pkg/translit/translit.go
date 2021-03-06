package translit

import (
	"strings"
	"unicode"

	"github.com/macroblock/imed/pkg/zlog/zlog"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var log = zlog.Instance("translit")

// Do -
func Do(text string) (string, error) {

	var (
		ruseng = map[rune]string{'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "c", 'ч': "ch", 'ш': "sh", 'щ': "sh", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya"}
		eng    = "abcdefghijklmnopqrstuvwxyz1234567890"

		ret, sep, underscore string
	)

	log.Debug("input string \"" + text + "\"")

	replaceY := runes.Map(func(r rune) rune {
		if r == 'й' {
			return 'y'
		}
		return r
	})

	t := transform.Chain(runes.ReplaceIllFormed(), norm.NFC, runes.Map(unicode.ToLower), replaceY, norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	src, _, err := transform.String(t, text)
	if err != nil {
		log.Error(err, "this point cannot be reached")
		return "", err
	}
	log.Debug("transform chain \"" + src + "\"")
	for _, r := range src {
		x, ok := ruseng[r]
		if !ok && strings.Index(eng, string(r)) != -1 {
			x = string(r)
			ok = true
		}
		if !ok {
			sep = underscore
		} else {
			ret += sep + x
			sep = ""
			underscore = "_"
		}
	}
	log.Debug("result \"" + ret + "\"")
	return ret, nil
}
