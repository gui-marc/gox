package compiler

type Node struct {
	Tag       string
	Attrs     []*Attr
	SelfClose string
	Children  []Node
	Close     *string
}

type Attr struct {
	Name  string
	Value string
}

type GOX struct {
	Nodes []*Node
}
