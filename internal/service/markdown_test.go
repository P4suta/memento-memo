package service

import (
	"strings"
	"testing"
)

func TestRenderHTML_Heading(t *testing.T) {
	html := RenderHTML("# Hello")
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("expected h1 tag, got: %s", html)
	}
}

func TestRenderHTML_List(t *testing.T) {
	html := RenderHTML("- item1\n- item2")
	if !strings.Contains(html, "<li>") {
		t.Errorf("expected li tags, got: %s", html)
	}
}

func TestRenderHTML_CodeBlock(t *testing.T) {
	html := RenderHTML("```go\nfmt.Println(\"hello\")\n```")
	if !strings.Contains(html, "<code") {
		t.Errorf("expected code tag, got: %s", html)
	}
}

func TestRenderHTML_GFM_Strikethrough(t *testing.T) {
	html := RenderHTML("~~deleted~~")
	if !strings.Contains(html, "<del>deleted</del>") {
		t.Errorf("expected del tag for GFM strikethrough, got: %s", html)
	}
}

func TestRenderHTML_GFM_Table(t *testing.T) {
	md := "| a | b |\n|---|---|\n| 1 | 2 |"
	html := RenderHTML(md)
	if !strings.Contains(html, "<table>") {
		t.Errorf("expected table tag, got: %s", html)
	}
}

func TestRenderHTML_GFM_Autolink(t *testing.T) {
	html := RenderHTML("Visit https://example.com")
	if !strings.Contains(html, `<a href="https://example.com"`) {
		t.Errorf("expected autolink, got: %s", html)
	}
}

func TestRenderHTML_XSS_ScriptTag(t *testing.T) {
	html := RenderHTML("<script>alert('xss')</script>")
	if strings.Contains(html, "<script>") {
		t.Errorf("XSS not sanitized: %s", html)
	}
}

func TestRenderHTML_XSS_OnEvent(t *testing.T) {
	html := RenderHTML(`<img src=x onerror="alert('xss')">`)
	if strings.Contains(html, "onerror") {
		t.Errorf("XSS event handler not sanitized: %s", html)
	}
}

func TestRenderHTML_EmptyString(t *testing.T) {
	html := RenderHTML("")
	if html != "" {
		t.Errorf("expected empty string for empty input, got: %q", html)
	}
}
