package ziptree

import (
	"bufio"
	"fmt"
	"io"
)

// FprintZipData converts zip binary contents to a string literal.
func FprintZipData(w io.Writer, zipData []byte) error {
	dest := bufio.NewWriter(w)
	for _, b := range zipData {
		if err := func(b byte) error {
			switch b {
			case '\n':
				_, err := dest.WriteString(`\n`)
				return err
			case '\\', '"':
				_, err := dest.WriteString(`\` + string(b))
				return err
			default:
				if (b >= 32 && b <= 126) || b == '\t' {
					return dest.WriteByte(b)
				}
				_, err := fmt.Fprintf(dest, "\\x%02x", b)
				return err
			}
		}(b); err != nil {
			return err
		}
	}
	return dest.Flush()
}
