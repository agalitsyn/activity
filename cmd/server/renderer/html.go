package renderer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
)

const (
	SmartBlock   = ""
	BaseBlock    = "base"
	ContentBlock = "content"
	ErrorBlock   = "error"
)

type pageData struct {
	// data for layout
	Path string

	// data for block
	Data any
}

type errorData struct {
	Message template.HTML
	Error   string
}

func newPageData(r *http.Request) pageData {
	return pageData{Path: r.URL.Path}
}

type HTMLRenderer struct {
	Debug            bool
	templateRenderer *TemplateRenderer
}

func NewHTMLRenderer(renderer *TemplateRenderer) *HTMLRenderer {
	return &HTMLRenderer{
		templateRenderer: renderer,
	}
}

func (c *HTMLRenderer) Render(w http.ResponseWriter, r *http.Request, status int, template, block string, data any) {
	logAttrs := []any{"template", template, "block", block}
	if block != SmartBlock && isHTMXRequest(r) {
		slog.Debug("render html", logAttrs...)
		c.templateRenderer.Render(w, status, template, block, data)
		return
	}

	if block == SmartBlock {
		block = BaseBlock
		if isHTMXRequest(r) {
			block = ContentBlock
		}
	}

	pd := newPageData(r)
	pd.Data = data

	c.templateRenderer.Render(w, status, template, block, pd)
}

func (c *HTMLRenderer) Error(w http.ResponseWriter, r *http.Request, status int, msg string, err error) {
	data := errorData{
		Message: template.HTML(msg),
	}
	if c.Debug && err != nil {
		data.Error = err.Error()
	}

	// Any HTMX errors are rendered as a static block on defined in template page region
	if isHTMXRequest(r) {
		w.Header().Add("HX-Retarget", "#general-error")
		w.Header().Add("HX-Reswap", "innerHTML")
		c.Render(w, r, http.StatusOK, "error.tmpl.html", ErrorBlock, data)
		return
	}

	if status == http.StatusNotFound {
		c.Render(w, r, status, "404.tmpl.html", BaseBlock, data)
		return
	}

	c.Render(w, r, status, "500.tmpl.html", BaseBlock, data)
}

func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

type TemplateRenderer struct {
	cache map[string]*template.Template
}

func NewTemplateRenderer(templatesCache map[string]*template.Template) *TemplateRenderer {
	return &TemplateRenderer{
		cache: templatesCache,
	}
}

func (s *TemplateRenderer) Render(w http.ResponseWriter, status int, template, block string, data any) {
	ts, ok := s.cache[template]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", template)
		slog.Error("could not fetch template", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, block, data)
	if err != nil {
		slog.Error("could not execute template", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		slog.Error("could not write template content", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func NewTemplateCache(
	embedFiles embed.FS,
	templatesFolder string,
	funcs template.FuncMap,
) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	root, err := fs.Glob(embedFiles, templatesFolder+"/*.tmpl.html")
	if err != nil {
		return nil, err
	}
	tree, err := fs.Glob(embedFiles, templatesFolder+"/*/*.tmpl.html")
	if err != nil {
		return nil, err
	}
	pages := append(root, tree...)

	for _, page := range pages {
		name := filepath.Base(page)
		patterns := []string{
			templatesFolder + "/*.tmpl.html",
			templatesFolder + "/partials/*.tmpl.html",
			page,
		}

		ts, err := template.New(name).Funcs(funcs).ParseFS(embedFiles, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
