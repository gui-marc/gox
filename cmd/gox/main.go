package main

import (
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
	parser := compiler.NewParticipleParser()

	input := `<div attr="value">Content</div>`
	ast, err := parser.ParseString("", input)
	if err != nil {
		return err
	}

	_ = ast

	return nil
}
