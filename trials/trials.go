package main

/*
IDEA 1
We should process templates as follows:
1) Use text/template Parse func to find all actions, save their types and location, remove them
// Note: how do we deal with conditionals and loops? How do we consider both branches, context-wise?
// Should we do the parsing twice, once with all conditionals and one without and check that the context after each cond is identical?
2) On the now stripped template we tokenize it as HTML, and do some fundamental rewrites (e.g. all attributes should now be dquoted)
While doing so we need to keep track of a translation matrix, since positions of the action nodes will now need to be updated
3) As we do so, we check whether we reached an action, and compute the context it appears in
5) Compose the correct escapers for actions
6) Put the Actions back in the template, with the escapers added to the action
7) Use text/template to render it


*/

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const tpl = `
{{.Action}}
<img src="/foo" foo='bar"' bar=lol/>
`

func main() {
	tokenization()
	parsing()
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func parsing() {
	fmt.Println("Parse") // This is way too destructive, I don't think we can use it for templates.
	nodes := must(html.ParseFragment(strings.NewReader(tpl), nil))
	for _, node := range nodes {
		check(html.Render(os.Stdout, node))
	}
}

func tokenization() {
	tknz := html.NewTokenizer(strings.NewReader(tpl))
	var output strings.Builder
loop:
	for {
		switch tkn := tknz.Next(); {
		case tkn == html.ErrorToken && errors.Is(tknz.Err(), io.EOF):
			break loop
		case tkn == html.ErrorToken:
			fmt.Printf("%v %q %v\n", tkn, string(tknz.Raw()), tknz.Err())
		case tkn == html.StartTagToken || tkn == html.SelfClosingTagToken:
			fmt.Printf("%v %q\n", tkn, string(tknz.Raw()))
			printTag(&output, tknz, tkn)
		default:
			output.Write(tknz.Raw())
			fmt.Printf("%v %q\n", tkn, string(tknz.Raw()))
		}
	}
	fmt.Println()
	fmt.Printf("Rewritten:\n%s", output.String())
}

func printTag(output *strings.Builder, tknz *html.Tokenizer, tkn html.TokenType) {
	// This doesn't tell us how the attributes were quoted, if they were quoted,
	// which makes it really hard to know how to potentially escape their values
	// if we impose double quotes.
	// We'd need to clunkily re-scan the raw string and guess which separator was
	// used, re-implementing some parsing logic.
	// Otherwise we just escape everything for double quotes since the tokenizer
	// returns unescaped strings.
	tag, hasAttr := tknz.TagName()
	output.WriteRune('<')
	fmt.Printf("↳ TagName: %q\n", string(tag))
	output.Write(tag)
	for hasAttr {
		key, val, more := tknz.TagAttr()
		fmt.Printf("↳ Attr: key:%q val:%q\n", string(key), string(val))
		output.WriteRune(' ')
		output.Write(key)
		output.WriteString(`="`)
		output.WriteString(html.EscapeString(string(val))) // We escape because we are forcing quoting
		output.WriteRune('"')
		hasAttr = more
	}
	if tkn == html.SelfClosingTagToken {
		output.WriteString("/>") // Problem: if last attribute is unquoted we parsed the closing slash as part of its value
	} else {
		output.WriteRune('>')
	}
}
