// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"bytes"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// absolutizeFeedHTML rewrites relative href/src/poster attributes in one entry's
// rendered content against the feed's own origin.
//
// Feed content is consumed outside the browser that fetched it, so a relative
// reference — local storage stores "/api/files/…" — has no base to resolve
// against and readers drop or break those images. mdUtil.MdToHTML already drops
// raw HTML, so re-rendering the fragment is safe.
func absolutizeFeedHTML(fragment []byte, base *url.URL) []byte {
	if base == nil || base.Host == "" || len(fragment) == 0 {
		return fragment
	}
	context := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(bytes.NewReader(fragment), context)
	if err != nil {
		return fragment
	}

	var out bytes.Buffer
	for _, node := range nodes {
		absolutizeNode(node, base)
		if err := html.Render(&out, node); err != nil {
			return fragment
		}
	}
	return out.Bytes()
}

func absolutizeNode(node *html.Node, base *url.URL) {
	if node.Type == html.ElementNode {
		for i := range node.Attr {
			attr := &node.Attr[i]
			switch attr.Key {
			case "href", "src", "poster":
				attr.Val = resolveFeedURL(base, attr.Val)
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		absolutizeNode(child, base)
	}
}

func resolveFeedURL(base *url.URL, raw string) string {
	if raw == "" || strings.HasPrefix(raw, "#") {
		return raw
	}
	if strings.HasPrefix(raw, "//") {
		return base.Scheme + ":" + raw
	}
	ref, err := url.Parse(raw)
	if err != nil || ref.IsAbs() {
		return raw
	}
	return base.ResolveReference(ref).String()
}
