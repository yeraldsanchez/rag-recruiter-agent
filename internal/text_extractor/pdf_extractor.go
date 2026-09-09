package extractor

import (
	"bytes"
	"io"
	"math"
	"regexp"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

type PDFExtractor struct {
}

func NewPDFExtractor() PDFExtractor {
	return PDFExtractor{}
}

func cleanText(raw string) string {
	texto := strings.Map(func(r rune) rune {
		if r == '\uFFFD' || r == 0 {
			return -1 // -1 elimina el carácter del string final
		}
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		return -1
	}, raw)
	reBullets := regexp.MustCompile(`[•\t]+`)
	texto = reBullets.ReplaceAllString(texto, " ")

	reCamel := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	texto = reCamel.ReplaceAllString(texto, "$1 $2")

	reSpaces := regexp.MustCompile(`\s+`)
	texto = reSpaces.ReplaceAllString(texto, " ")

	return strings.TrimSpace(texto)
}

func (p PDFExtractor) ExtractText(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	bytesReader := bytes.NewReader(data)

	pdfReader, err := pdf.NewReader(bytesReader, int64(len(data)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	for pageIndex := 1; pageIndex <= pdfReader.NumPage(); pageIndex++ {
		page := pdfReader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		content := page.Content()
		var words []pdf.Text

		for _, entry := range content.Text {
			words = append(words, entry)
		}

		if len(words) == 0 {
			continue
		}
		for i := 0; i < len(words); i++ {
			curr := words[i]
			buf.WriteString(curr.S)

			if i < len(words)-1 {
				next := words[i+1]

				deltaY := math.Abs(curr.Y - next.Y)
				if deltaY > 3.0 {
					buf.WriteString("\n")
					continue
				}

				deltaX := next.X - (curr.X + curr.W)
				if deltaX > 1.5 || (deltaX < 0 && math.Abs(deltaX) > 10) {
					buf.WriteString(" ")
				}
			}
		}
		buf.WriteString("\n")
	}
	return cleanText(buf.String()), nil
}
