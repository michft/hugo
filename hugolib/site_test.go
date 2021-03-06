package hugolib

import (
	"bytes"
	"fmt"
	"github.com/spf13/hugo/source"
	"html/template"
	"strings"
	"testing"
)

const (
	TEMPLATE_TITLE    = "{{ .Title }}"
	PAGE_SIMPLE_TITLE = `---
title: simple template
---
content`

	TEMPLATE_MISSING_FUNC        = "{{ .Title | funcdoesnotexists }}"
	TEMPLATE_FUNC                = "{{ .Title | urlize }}"
	TEMPLATE_CONTENT             = "{{ .Content }}"
	TEMPLATE_DATE                = "{{ .Date }}"
	INVALID_TEMPLATE_FORMAT_DATE = "{{ .Date.Format time.RFC3339 }}"
	TEMPLATE_WITH_URL            = "<a href=\"foobar.jpg\">Going</a>"
	PAGE_URL_SPECIFIED           = `---
title: simple template
url: "mycategory/my-whatever-content/"
---
content`

	PAGE_WITH_MD = `---
title: page with md
---
# heading 1
text
## heading 2
more text
`
)

func pageMust(p *Page, err error) *Page {
	if err != nil {
		panic(err)
	}
	return p
}

func TestDegenerateRenderThingMissingTemplate(t *testing.T) {
	p, _ := ReadFrom(strings.NewReader(PAGE_SIMPLE_TITLE), "content/a/file.md")
	s := new(Site)
	s.prepTemplates()
	_, err := s.RenderThing(p, "foobar")
	if err == nil {
		t.Errorf("Expected err to be returned when missing the template.")
	}
}

func TestAddInvalidTemplate(t *testing.T) {
	s := new(Site)
	s.prepTemplates()
	err := s.addTemplate("missing", TEMPLATE_MISSING_FUNC)
	if err == nil {
		t.Fatalf("Expecting the template to return an error")
	}
}

func matchRender(t *testing.T, s *Site, p *Page, tmplName string, expected string) {
	content, err := s.RenderThing(p, tmplName)
	if err != nil {
		t.Fatalf("Unable to render template.")
	}

	if string(content.Bytes()) != expected {
		t.Fatalf("Content did not match expected: %s. got: %s", expected, content)
	}
}

func TestRenderThing(t *testing.T) {
	tests := []struct {
		content  string
		template string
		expected string
	}{
		{PAGE_SIMPLE_TITLE, TEMPLATE_TITLE, "simple template"},
		{PAGE_SIMPLE_TITLE, TEMPLATE_FUNC, "simple-template"},
		{PAGE_WITH_MD, TEMPLATE_CONTENT, "<h1>heading 1</h1>\n\n<p>text</p>\n\n<h2>heading 2</h2>\n\n<p>more text</p>\n"},
		{SIMPLE_PAGE_RFC3339_DATE, TEMPLATE_DATE, "2013-05-17 16:59:30 &#43;0000 UTC"},
	}

	s := new(Site)
	s.prepTemplates()

	for i, test := range tests {
		p, err := ReadFrom(strings.NewReader(test.content), "content/a/file.md")
		if err != nil {
			t.Fatalf("Error parsing buffer: %s", err)
		}
		templateName := fmt.Sprintf("foobar%d", i)
		err = s.addTemplate(templateName, test.template)
		if err != nil {
			t.Fatalf("Unable to add template")
		}

		p.Content = template.HTML(p.Content)
		html, err2 := s.RenderThing(p, templateName)
		if err2 != nil {
			t.Errorf("Unable to render html: %s", err)
		}

		if string(html.Bytes()) != test.expected {
			t.Errorf("Content does not match.\nExpected\n\t'%q'\ngot\n\t'%q'", test.expected, html)
		}
	}
}

func TestRenderThingOrDefault(t *testing.T) {
	tests := []struct {
		content  string
		missing  bool
		template string
		expected string
	}{
		{PAGE_SIMPLE_TITLE, true, TEMPLATE_TITLE, "simple template"},
		{PAGE_SIMPLE_TITLE, true, TEMPLATE_FUNC, "simple-template"},
		{PAGE_SIMPLE_TITLE, false, TEMPLATE_TITLE, "simple template"},
		{PAGE_SIMPLE_TITLE, false, TEMPLATE_FUNC, "simple-template"},
	}

	s := new(Site)
	s.prepTemplates()

	for i, test := range tests {
		p, err := ReadFrom(strings.NewReader(PAGE_SIMPLE_TITLE), "content/a/file.md")
		if err != nil {
			t.Fatalf("Error parsing buffer: %s", err)
		}
		templateName := fmt.Sprintf("default%d", i)
		err = s.addTemplate(templateName, test.template)
		if err != nil {
			t.Fatalf("Unable to add template")
		}

		var html *bytes.Buffer
		var err2 error
		if test.missing {
			html, err2 = s.RenderThingOrDefault(p, "missing", templateName)
		} else {
			html, err2 = s.RenderThingOrDefault(p, templateName, "missing_default")
		}

		if err2 != nil {
			t.Errorf("Unable to render html: %s", err)
		}

		if string(html.Bytes()) != test.expected {
			t.Errorf("Content does not match.  Expected '%s', got '%s'", test.expected, html)
		}
	}
}

func TestTargetPath(t *testing.T) {
	tests := []struct {
		doc             string
		content         string
		expectedOutFile string
		expectedSection string
	}{
		{"content/a/file.md", PAGE_URL_SPECIFIED, "mycategory/my-whatever-content/index.html", "a"},
		{"content/b/file.md", SIMPLE_PAGE, "b/file.html", "b"},
		{"a/file.md", SIMPLE_PAGE, "a/file.html", "a"},
		{"file.md", SIMPLE_PAGE, "file.html", ""},
	}

	if true {
		return
	}
	for _, test := range tests {
		s := &Site{
			Config: Config{ContentDir: "content"},
		}
		p := pageMust(ReadFrom(strings.NewReader(test.content), s.Config.GetAbsPath(test.doc)))

		expected := test.expectedOutFile

		if p.TargetPath() != expected {
			t.Errorf("%s => OutFile  expected: '%s', got: '%s'", test.doc, expected, p.TargetPath())
		}

		if p.Section != test.expectedSection {
			t.Errorf("%s => p.Section expected: %s, got: %s", test.doc, test.expectedSection, p.Section)
		}
	}
}

func TestSkipRender(t *testing.T) {
	files := make(map[string][]byte)
	target := &InMemoryTarget{files: files}
	sources := []source.ByteSource{
		{"sect/doc1.html", []byte("---\nmarkup: markdown\n---\n# title\nsome *content*"), "sect"},
		{"sect/doc2.html", []byte("<!doctype html><html><body>more content</body></html>"), "sect"},
		{"sect/doc3.md", []byte("# doc3\n*some* content"), "sect"},
		{"sect/doc4.md", []byte("---\ntitle: doc4\n---\n# doc4\n*some content*"), "sect"},
		{"sect/doc5.html", []byte("<!doctype html><html>{{ template \"head\" }}<body>body5</body></html>"), "sect"},
		{"doc7.html", []byte("<html><body>doc7 content</body></html>"), ""},
	}

	s := &Site{
		Target: target,
		Config: Config{BaseUrl: "http://auth/bub/"},
		Source: &source.InMemorySource{sources},
	}
	s.initializeSiteInfo()
	s.prepTemplates()

	must(s.addTemplate("_default/single.html", "{{.Content}}"))
	must(s.addTemplate("head", "<head><script src=\"script.js\"></script></head>"))

	if err := s.CreatePages(); err != nil {
		t.Fatalf("Unable to create pages: %s", err)
	}

	if err := s.BuildSiteMeta(); err != nil {
		t.Fatalf("Unable to build site metadata: %s", err)
	}

	if err := s.RenderPages(); err != nil {
		t.Fatalf("Unable to render pages. %s", err)
	}

	tests := []struct {
		doc      string
		expected string
	}{
		{"sect/doc1.html", "<html><head></head><body><h1>title</h1>\n\n<p>some <em>content</em></p>\n</body></html>"},
		{"sect/doc2.html", "<!DOCTYPE html><html><head></head><body>more content</body></html>"},
		{"sect/doc3.html", "<html><head></head><body><h1>doc3</h1>\n\n<p><em>some</em> content</p>\n</body></html>"},
		{"sect/doc4.html", "<html><head></head><body><h1>doc4</h1>\n\n<p><em>some content</em></p>\n</body></html>"},
		{"sect/doc5.html", "<!DOCTYPE html><html><head><script src=\"http://auth/bub/script.js\"></script></head><body>body5</body></html>"},
		{"doc7.html", "<html><head></head><body>doc7 content</body></html>"},
	}

	for _, test := range tests {
		content, ok := target.files[test.doc]
		if !ok {
			t.Fatalf("Did not find %s in target. %v", test.doc, target.files)
		}

		if !bytes.Equal(content, []byte(test.expected)) {
			t.Errorf("%s content expected:\n%q\ngot:\n%q", test.doc, test.expected, string(content))
		}
	}
}

func TestAbsUrlify(t *testing.T) {
	files := make(map[string][]byte)
	target := &InMemoryTarget{files: files}
	sources := []source.ByteSource{
		{"sect/doc1.html", []byte("<!doctype html><html><head></head><body><a href=\"#frag1\">link</a></body></html>"), "sect"},
		{"content/blue/doc2.html", []byte("---\nf: t\n---\n<!doctype html><html><body>more content</body></html>"), "blue"},
	}
	s := &Site{
		Target: target,
		Config: Config{BaseUrl: "http://auth/bub/"},
		Source: &source.InMemorySource{sources},
	}
	s.initializeSiteInfo()
	s.prepTemplates()
	must(s.addTemplate("blue/single.html", TEMPLATE_WITH_URL))

	if err := s.CreatePages(); err != nil {
		t.Fatalf("Unable to create pages: %s", err)
	}

	if err := s.BuildSiteMeta(); err != nil {
		t.Fatalf("Unable to build site metadata: %s", err)
	}

	if err := s.RenderPages(); err != nil {
		t.Fatalf("Unable to render pages. %s", err)
	}

	tests := []struct {
		file, expected string
	}{
		{"content/blue/doc2.html", "<html><head></head><body><a href=\"http://auth/bub/foobar.jpg\">Going</a></body></html>"},
		{"sect/doc1.html", "<!DOCTYPE html><html><head></head><body><a href=\"#frag1\">link</a></body></html>"},
	}

	for _, test := range tests {
		content, ok := target.files[test.file]
		if !ok {
			t.Fatalf("Unable to locate rendered content: %s", test.file)
		}

		expected := test.expected
		if string(content) != expected {
			t.Errorf("AbsUrlify content expected:\n%q\ngot\n%q", expected, string(content))
		}
	}
}
