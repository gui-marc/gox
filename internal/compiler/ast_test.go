package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAst(t *testing.T) {
	divCloseTag := "div"
	spanCloseTag := "span"

	testCases := []struct {
		name          string
		input         string
		result        *GOX
		expectedError error
	}{
		{
			name:  "it should see a normal tag",
			input: `<div></div>`,
			result: &GOX{
				Nodes: []*Node{
					{Tag: "div", SelfClose: ">", Close: &divCloseTag},
				},
			},
		},
		{
			name:  "it should see a self-closing tag",
			input: `<img />`,
			result: &GOX{
				Nodes: []*Node{
					{Tag: "img", SelfClose: "/"},
				},
			},
		},
		{
			name:  "it should see a self-closing tag with attributes",
			input: `<img src="image.png" alt="Image" />`,
			result: &GOX{
				Nodes: []*Node{
					{Tag: "img", SelfClose: "/", Attrs: []*Attr{
						{Name: "src", Value: `"image.png"`},
						{Name: "alt", Value: `"Image"`},
					}},
				},
			},
		},
		{
			name:  "it should see a normal tag with spaces",
			input: `<div  ></div >`,
			result: &GOX{
				Nodes: []*Node{
					{Tag: "div", SelfClose: ">", Close: &divCloseTag},
				},
			},
		},
		{
			name:  "it should see a tag with attributes",
			input: `<div class="test"></div>`,
			result: &GOX{
				Nodes: []*Node{
					{

						Tag:       "div",
						Attrs:     []*Attr{{Name: "class", Value: `"test"`}},
						SelfClose: ">",
						Close:     &divCloseTag,
					},
				},
			},
		},
		{
			name:  "it should see a tag with multiple attributes",
			input: `<div class="test" id="main"></div>`,
			result: &GOX{
				Nodes: []*Node{
					{

						Tag: "div",
						Attrs: []*Attr{
							{Name: "class", Value: `"test"`},
							{Name: "id", Value: `"main"`},
						},
						SelfClose: ">",
						Close:     &divCloseTag,
					},
				},
			},
		},
		{
			name:  "it should work with nested nodes",
			input: `<div><span></span></div>`,
			result: &GOX{
				Nodes: []*Node{
					{
						Tag: "div",
						Children: []Node{
							{
								Tag:       "span",
								SelfClose: ">",
								Close:     &spanCloseTag,
							},
						},
						SelfClose: ">",
						Close:     &divCloseTag,
					},
				},
			},
		},
		{
			name:  "it should work with nested nodes and self-closing tags",
			input: `<div><span/></div>`,
			result: &GOX{
				Nodes: []*Node{
					{
						Tag: "div",
						Children: []Node{
							{
								Tag:       "span",
								SelfClose: "/",
								Close:     nil,
							},
						},
						Close:     &divCloseTag,
						SelfClose: ">",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			parser := NewParticipleParser()

			ast, err := parser.ParseString("", tc.input)
			if err != nil {
				if tc.expectedError != nil {
					require.Equal(t, tc.expectedError, err)

					return
				}

				t.Fatal(err)
			}

			require.Equal(t, tc.result, ast)
		})
	}
}
