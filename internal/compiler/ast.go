package compiler

type Node struct {
	Tag       string  `"<" @Ident`
	Attrs     []*Attr `@@*`
	SelfClose string  `(@"/" ">" | @">")`
	Children  []Node  `@@*`
	Close     *string `  ( "<" "/" @Ident ">" )?`
}

type CloseTag struct {
	Tag string `"<" "/" @Ident ">" `
}

type Attr struct {
	Name  string `@Ident "="`
	Value string `@String`
}

type GOX struct {
	Nodes []*Node `@@*`
}
