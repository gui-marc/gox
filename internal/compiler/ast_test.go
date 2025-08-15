package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAst(t *testing.T) {
	divCloseTag := "div"
	spanCloseTag := "span"

	testCases := []struct {
		name  string
		input string
		want  *GOX
	}{
		{
			name:  "normal tag",
			input: `<div></div>`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "div",
						SelfClose: ">",
						Attrs:     []*Attr{},
						Children:  []Node{},
						Close:     &divCloseTag,
					},
				},
			},
		},
		{
			name:  "self-closing tag",
			input: `<img />`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "img",
						Attrs:     []*Attr{},
						SelfClose: "/",
					},
				},
			},
		},
		{
			name:  "self-closing tag with attributes",
			input: `<img src="image.png" alt="Image" />`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "img",
						SelfClose: "/",
						Attrs: []*Attr{
							{Name: "src", Value: `"image.png"`},
							{Name: "alt", Value: `"Image"`},
						},
					},
				},
			},
		},
		{
			name:  "normal tag with spaces",
			input: `<div  ></div >`,
			want: &GOX{
				Nodes: []*Node{
					{
						Attrs:     []*Attr{},
						Children:  []Node{},
						Tag:       "div",
						SelfClose: ">",
						Close:     &divCloseTag,
					},
				},
			},
		},
		{
			name:  "tag with single attribute",
			input: `<div class="test"></div>`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "div",
						SelfClose: ">",
						Attrs:     []*Attr{{Name: "class", Value: `"test"`}},
						Children:  []Node{},
						Close:     &divCloseTag,
					},
				},
			},
		},
		{
			name:  "tag with multiple attributes",
			input: `<div class="test" id="main"></div>`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "div",
						SelfClose: ">",
						Attrs: []*Attr{
							{Name: "class", Value: `"test"`},
							{Name: "id", Value: `"main"`},
						},
						Children: []Node{},
						Close:    &divCloseTag,
					},
				},
			},
		},
		{
			name:  "nested nodes",
			input: `<div><span></span></div>`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "div",
						SelfClose: ">",
						Attrs:     []*Attr{},
						Children: []Node{
							{
								Tag:       "span",
								SelfClose: ">",
								Attrs:     []*Attr{},
								Children:  []Node{},
								Close:     &spanCloseTag,
							},
						},
						Close: &divCloseTag,
					},
				},
			},
		},
		{
			name:  "nested nodes with self-closing tag",
			input: `<div><span/></div>`,
			want: &GOX{
				Nodes: []*Node{
					{
						Tag:       "div",
						Attrs:     []*Attr{},
						SelfClose: ">",
						Children: []Node{
							{
								Tag:       "span",
								Attrs:     []*Attr{},
								SelfClose: "/",
							},
						},
						Close: &divCloseTag,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser(tc.input)
			got := parser.Parse()
			require.Equal(t, tc.want, got)
		})
	}
}
