// goxc/parser_test.go
package main

import (
	"go/format"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser_Success(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "it should handle single-line return without parenthesis",
			input:    `func a() { return <div>Hello</div> }`,
			expected: `func a() { return gox.El("div", gox.Props{}, gox.Text("Hello")) }`,
		},
		{
			name:     "it should handle variable assignment without parenthesis",
			input:    `func a() { myVar := <p>text</p> }`,
			expected: `func a() { myVar := gox.El("p", gox.Props{}, gox.Text("text")) }`,
		},
		{
			name: "it should handle multi-line block with parenthesis but not render them",
			input: `func a() {
				return (
					<main>
						<h1>Title</h1>
					</main>
				)
			}`,
			expected: `func a() gox.Node {
				return gox.El("main", gox.Props{}, gox.El("h1", gox.Props{}, gox.Text("Title")))
			}`,
		},
		{
			name:     "it should parse a simple self-closing element",
			input:    `var n = <hr />`,
			expected: `var n = gox.El("hr", gox.Props{})`,
		},
		{
			name:     "it should parse a simple element with text content",
			input:    `var n = <div>Hello World</div>`,
			expected: `var n = gox.El("div", gox.Props{}, gox.Text("Hello World"))`,
		},
		{
			name:     "it should handle multiple children of different types",
			input:    `var n = <div><h1>Title</h1><p>Hello, {user.Name}</p><hr /></div>`,
			expected: `var n = gox.El("div", gox.Props{}, gox.El("h1", gox.Props{}, gox.Text("Title")), gox.El("p", gox.Props{}, gox.Text("Hello, "), gox.Text(user.Name)), gox.El("hr", gox.Props{}))`,
		},
		{
			name: "it should maintain standard go expressions",
			input: `
				func calculate() int {
					if 5 < 10 {
						return 5
					}
					return 10
				}
			`,
			expected: `
				func calculate() int {
					if 5 < 10 {
						return 5
					}
					return 10
				}
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			p := NewParser([]byte(tc.input))
			result, err := p.Parse()
			require.NoError(t, err)

			formattedResult, err := format.Source(result)
			require.NoError(t, err, "Resulting code is not valid Go")

			formattedExpected, err := format.Source([]byte(tc.expected))
			require.NoError(t, err, "Expected code is not valid Go")

			require.Equal(t, string(formattedExpected), string(formattedResult))
		})
	}
}
