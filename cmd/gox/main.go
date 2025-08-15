package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/gui-marc/gox/internal/compiler"
)

type CLI struct {
	Input string ``
}

func main() {
	var cli CLI
	kong.Parse(&cli)

	if err := cli.Run(); err != nil {
		panic(err)
	}
}

func (c *CLI) Run() error {
	input := `<div attr="value">Content</div>`
	parser := compiler.NewParser(input)

	parsed := parser.Parse()

	fmt.Println(parsed)

	return nil
}
