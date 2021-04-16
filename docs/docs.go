package docs

import (
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type FunctionType string

const (
	FnNormal FunctionType = "normal"
	FnMethod FunctionType = "method"
)

var (
	reType   = regexp.MustCompile(`^type\s([a-zA-Z0-9_]+)\s([a-zA-Z0-9_]+).*`)
	reFunc   = regexp.MustCompile(`^func\s([a-zA-Z0-9_]+)\(.*\).*$`)
	reMethod = regexp.MustCompile(`^func\s\(([a-zA-Z0-9\*\s]+)\)\s([a-zA-Z0-9]+).+$`)
)

const BASE = "https://pkg.go.dev/"

type Function struct {
	Name      string       `json:"name"`
	Type      FunctionType `json:"type"`
	Signature string       `json:"signature"`
	MethodOf  string       `json:"methodOf"`

	Example  string   `json:"example"`
	Comments []string `json:"comments"`
}

type Type struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Signature string `json:"signature"`

	Comments []string `json:"comments"`
}

type Doc struct {
	URL       string     `json:"url"`
	Name      string     `json:"name"`
	Overview  string     `json:"overview"`
	Types     []Type     `json:"types"`
	Functions []Function `json:"functions"`
}

// GetDoc returns a document representing the specified package/module.
func GetDoc(pkg string) (*Doc, error) {
	resp, err := http.Get(BASE + pkg)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	var (
		funcs    []Function
		types    []Type
		overview string

		sign string
		par  string
	)

	// funcs
	doc.Find("div.Documentation-function").Each(func(_ int, item *goquery.Selection) {
		sign = item.Find("pre").First().Text()
		sign = strings.ReplaceAll(sign, "\n", "")
		fn := Function{
			Signature: sign,
			Type:      FnNormal,
		}
		if matches := reFunc.FindStringSubmatch(sign); len(matches) == 2 {
			fn.Name = matches[1]
		} else {
			return
		}
		fn.Example = item.Find("textarea.Documentation-exampleCode").First().Text()
		item.Find("p").Each(func(_ int, p *goquery.Selection) {
			par = p.Text()
			if par != "" {
				fn.Comments = append(fn.Comments, par)
			}
		})
		funcs = append(funcs, fn)
	})

	// type funcs
	doc.Find("div.Documentation-typeFunc").Each(func(_ int, item *goquery.Selection) {
		sign = item.Find("pre").First().Text()
		sign = strings.ReplaceAll(sign, "\n", "")
		fn := Function{
			Signature: sign,
			Type:      FnNormal,
		}
		if matches := reFunc.FindStringSubmatch(sign); len(matches) == 2 {
			fn.Name = matches[1]
		} else {
			return
		}
		fn.Example = item.Find("textarea.Documentation-exampleCode").First().Text()
		item.Find("p").Each(func(_ int, p *goquery.Selection) {
			par = p.Text()
			if par != "" {
				fn.Comments = append(fn.Comments, par)
			}
		})
		funcs = append(funcs, fn)
	})

	// methods
	doc.Find("div.Documentation-typeMethod").Each(func(_ int, item *goquery.Selection) {
		sign = item.Find("pre").First().Text()
		sign = strings.ReplaceAll(sign, "\n", "")
		fn := Function{
			Signature: sign,
			Type:      FnMethod,
		}
		if matches := reMethod.FindStringSubmatch(sign); len(matches) == 3 {
			fn.MethodOf = extractType(matches[1])
			fn.Name = matches[2]
		} else {
			return
		}

		fn.Example = item.Find("textarea.Documentation-exampleCode").First().Text()
		item.Find("p").Each(func(_ int, p *goquery.Selection) {
			par = p.Text()
			if par != "" {
				fn.Comments = append(fn.Comments, par)
			}
		})
		funcs = append(funcs, fn)
	})

	// types
	doc.Find("div.Documentation-type").Each(func(_ int, item *goquery.Selection) {
		sign = item.Find("pre").First().Text()
		//sign = strings.ReplaceAll(sign, "\n", "")
		t := Type{Signature: sign}
		if matches := reType.FindStringSubmatch(sign); len(matches) == 3 {
			t.Name = matches[1]
			t.Type = matches[2]
		} else {
			return
		}
		item.Find("p").Each(func(_ int, p *goquery.Selection) {
			par = p.Text()
			if par != "" {
				t.Comments = append(t.Comments, par)
			}
		})
		types = append(types, t)
	})

	// overview
	doc.Find("section.Documentation-overview > p").Each(func(_ int, p *goquery.Selection) {
		par = p.Text()
		if par != "" {
			overview += par + "\n"
		}
	})

	return &Doc{
		URL:       BASE + pkg,
		Overview:  overview,
		Name:      pkg,
		Functions: funcs,
		Types:     types,
	}, nil
}

// extractType extracts the type from a method definition
// i.e, `t *Type` -> `Type`
func extractType(s string) string {

	for i := len(s) - 1; i >= 0; i-- {
		if unicode.IsSpace(rune(s[i])) ||
			s[i] == '*' {
			return s[i+1:]
		}
	}
	return s
}
