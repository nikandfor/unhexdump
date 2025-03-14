package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const digits = "0123456789abcdef"

var (
	format  = flag.String("f", "hex-array", "format: hex, hex-array, decimal-array, binary")
	comment = flag.Bool("comment", false, "add comment with offsets to arrays output")
	line    = flag.Int("w", 30, "width")
)

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin")
	}

	r := make([]byte, len(b)/2)
	w := 0

	var x uint64
	var bs int
	i := 0

	for i < len(b) {
		// offset
		i = skipNum(b, i)

		for {
			i = skipSpaces(b, i)
			if i == len(b) || b[i] == '|' || b[i] == '\n' {
				break
			}

			x, bs, i = parseNum(b, i)
			if bs == 0 {
				return fmt.Errorf("unexpected symbol: %q at 0x%x", b[i], i)
			}

			w = encodeNum(r, w, x, bs)
		}

		i = skipLine(b, i)
	}

	r = r[:w]

	switch *format {
	case "hex":
		fmt.Printf("% x\n", r)
	case "hex-array":
		for i := 0; i < w; i += *line {
			if *comment {
				r = append(r, '/', '*', ' ', digits[(i>>12)&0xf], digits[(i>>8)&0xf], digits[(i>>4)&0xf], digits[(i)&0xf], ':', ' ', '*', '/')
			}

			for j := i; j < i+*line && j < w; j++ {
				r = append(r, ' ', '0', 'x', digits[r[j]>>4], digits[r[j]&0xf], ',')
			}

			r = append(r, '\n')
		}

		fmt.Printf("%s\n", r[w:])
	case "decimal-array":
		for i := 0; i < w; i += *line {
			if *comment {
				r = append(r, '/', '*', ' ', digits[(i>>12)&0xf], digits[(i>>8)&0xf], digits[(i>>4)&0xf], digits[(i)&0xf], ':', ' ', '*', '/')
			}

			for j := i; j < i+*line && j < w; j++ {
				r = append(r, ' ')

				b := r[j]

				if b >= 100 {
					r = append(r, digits[b/100])
				}
				b %= 100

				if b >= 10 {
					r = append(r, digits[b/10])
				}
				b %= 10

				r = append(r, digits[b], ',')
			}

			r = append(r, '\n')
		}

		fmt.Printf("%s\n", r[w:])
	case "binary":
		fmt.Printf("%s", r)
	default:
		return fmt.Errorf("unsupported format: %v", *format)
	}

	return nil
}

func encodeNum(b []byte, i int, x uint64, bs int) int {
	for j := bs - 1; j >= 0; j-- {
		bb := byte(x >> (8 * j))
		b[i] = bb
		i++
	}

	return i
}

func parseNum(b []byte, i int) (x uint64, bs, _ int) {
loop:
	for i < len(b) {
		switch {
		case b[i] >= '0' && b[i] <= '9':
			x = x<<4 | uint64(b[i]-'0')
		case b[i] >= 'a' && b[i] <= 'f':
			x = x<<4 | uint64(b[i]-'a'+10)
		case b[i] >= 'A' && b[i] <= 'F':
			x = x<<4 | uint64(b[i]-'A'+10)
		default:
			break loop
		}

		i++
		bs++
	}

	bs /= 2

	return x, bs, i
}

func skipNum(b []byte, i int) int {
	for i < len(b) && (b[i] >= '0' && b[i] <= '9' || b[i] >= 'a' && b[i] <= 'f' || b[i] >= 'A' && b[i] <= 'F') {
		i++
	}

	return i
}

func skipSpaces(b []byte, i int) int {
	for i < len(b) && b[i] == ' ' {
		i++
	}

	return i
}

func skipLine(b []byte, i int) int {
	for i < len(b) && b[i] != '\n' {
		i++
	}

	if i != len(b) && b[i] == '\n' {
		i++
	}

	return i
}
