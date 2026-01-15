package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"dnsmesh/pkg/crypto"
)

func main() {
	cipher := flag.String("cipher", "", "base64 ciphertext to decrypt")
	trim := flag.Bool("trim", true, "trim whitespace from input")
	flag.Parse()

	input := *cipher
	if input == "" && flag.NArg() > 0 {
		input = flag.Arg(0)
	}
	if input == "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read stdin:", err)
			os.Exit(1)
		}
		input = string(data)
	}
	if *trim {
		input = strings.TrimSpace(input)
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "missing ciphertext (use -cipher, arg, or stdin)")
		os.Exit(1)
	}

	if err := crypto.Initialize(); err != nil {
		fmt.Fprintln(os.Stderr, "init encryption key:", err)
		os.Exit(1)
	}

	plaintext, err := crypto.Decrypt(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "decrypt:", err)
		os.Exit(1)
	}

	fmt.Print(plaintext)
}
