package gox_test

import (
	"bytes"
	"testing"

	"maps"

	"github.com/gui-marc/gox"
	"github.com/stretchr/testify/require"
)

// --- Test Components ---

func Card(props gox.Props, children ...gox.Node) gox.Node {
	original := gox.Props{
		"class": "card",
	}

	maps.Copy(original, props)

	return gox.El("div", original, children...)
}

func UserProfile(props gox.Props, children ...gox.Node) gox.Node {
	name := "Guest"
	if n, ok := props["name"].(string); ok {
		name = n
	}
	return gox.El("p", nil, gox.Text("User: "), gox.Text(name))
}

func PageLayout(props gox.Props, children ...gox.Node) gox.Node {
	return Card(gox.Props{"id": "layout"}, children...)
}

// --- Test Functions ---

func TestRenderTextNode(t *testing.T) {
	testCases := []struct {
		name   string
		input  any
		output string
	}{
		{
			name:   "it should render a simple text",
			input:  "Hello, World!",
			output: "Hello, World!",
		},
		{
			name:   "it should escape HTML characters",
			input:  "<script>alert('XSS')</script>",
			output: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:   "it should render an empty string",
			input:  "",
			output: "",
		},
		{
			name:   "it should render an integer as text",
			input:  12345,
			output: "12345",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			text := gox.Text(tc.input)

			var buf bytes.Buffer
			err := text.Render(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.output, buf.String())
		})
	}
}

func TestRenderElementNode(t *testing.T) {
	testCases := []struct {
		name   string
		input  gox.Node
		output string
	}{
		{
			name:   "it should render a simple div",
			input:  gox.El("div", nil, gox.Text("Hello, World!")),
			output: `<div>Hello, World!</div>`,
		},
		{
			name:   "it should render an element with no children",
			input:  gox.El("hr", nil),
			output: `<hr></hr>`,
		},
		{
			name:   "it should render an element with no props",
			input:  gox.El("p", nil, gox.Text("No props here.")),
			output: `<p>No props here.</p>`,
		},
		{
			name:   "it should render an element with various prop types",
			input:  gox.El("input", gox.Props{"type": "checkbox", "data-id": 123, "checked": true}),
			output: `<input type="checkbox" data-id="123" checked="true"></input>`,
		},
		{
			name: "it should render a nested structure",
			input: gox.El("div", gox.Props{"class": "container"},
				gox.Text("Hello, "),
				gox.El("span", gox.Props{"class": "highlight"}, gox.Text("World!")),
			),
			output: `<div class="container">Hello, <span class="highlight">World!</span></div>`,
		},
		{
			name: "it should render deeply nested elements",
			input: gox.El("main", nil,
				gox.El("section", gox.Props{"id": "first"},
					gox.El("p", nil, gox.Text("Paragraph 1")),
				),
			),
			output: `<main><section id="first"><p>Paragraph 1</p></section></main>`,
		},
		{
			name:   "it should render an element with multiple children of mixed types",
			input:  gox.El("ul", nil, gox.Text("List:"), gox.El("li", nil, gox.Text("one")), gox.El("li", nil, gox.Text("two"))),
			output: `<ul>List:<li>one</li><li>two</li></ul>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			var buf bytes.Buffer
			err := tc.input.Render(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.output, buf.String())
		})
	}
}

func TestRenderComponentNode(t *testing.T) {
	testCases := []struct {
		name   string
		input  gox.Node
		output string
	}{
		{
			name:   "it should render a simple component with no props or children",
			input:  gox.Component(UserProfile, nil),
			output: `<p>User: Guest</p>`,
		},
		{
			name:   "it should render a component with props",
			input:  gox.Component(UserProfile, gox.Props{"name": "Alice"}),
			output: `<p>User: Alice</p>`,
		},
		{
			name: "it should render a component that wraps a single child",
			input: gox.Component(Card, nil,
				gox.Text("Child content"),
			),
			output: `<div class="card">Child content</div>`,
		},
		{
			name: "it should render a component that wraps multiple children",
			input: gox.Component(Card, nil,
				gox.El("h1", nil, gox.Text("Title")),
				gox.El("p", nil, gox.Text("Description")),
			),
			output: `<div class="card"><h1>Title</h1><p>Description</p></div>`,
		},
		{
			name: "it should render a component that returns another component",
			input: gox.Component(PageLayout, nil,
				gox.Text("Page content"),
			),
			output: `<div class="card" id="layout">Page content</div>`,
		},
		{
			name:   "it should render a component that returns nil",
			input:  gox.Component(func(props gox.Props, children ...gox.Node) gox.Node { return nil }, nil),
			output: ``,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			var buf bytes.Buffer
			err := tc.input.Render(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.output, buf.String())
		})
	}
}

func TestRenderRootNode(t *testing.T) {
	testCases := []struct {
		name   string
		input  gox.Node
		output string
	}{
		{
			name:   "it should render a simple div node",
			input:  gox.El("div", nil, gox.Text("Hello, World!")),
			output: `<div>Hello, World!</div>`,
		},
		{
			name:   "it should render a div node with props",
			input:  gox.El("div", gox.Props{"class": "container"}, gox.Text("Hello, World!")),
			output: `<div class="container">Hello, World!</div>`,
		},
		{
			name:   "it should render text as root node",
			input:  gox.Text("just some plain text"),
			output: `just some plain text`,
		},
		{
			name:   "it should render a component as root node",
			input:  gox.Component(UserProfile, gox.Props{"name": "Root"}),
			output: `<p>User: Root</p>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			result, err := gox.Render(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.output, result)
		})
	}
}
