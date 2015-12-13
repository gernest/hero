package hero

import (
	"html/template"
	"io"

	"github.com/a8m/mark"
	"github.com/gernest/hot"
)

var viewFuncs = template.FuncMap{
	"markdown": toMarkdown,
}

// View is a interface for templates rendering
type View interface {
	Render(out io.Writer, templateName string, data interface{}) error
}

// DefaultView implements View interface
type DefaultView struct {
	tpl *hot.Template
}

// NewDefaultView returns a a DefaultView instance with all templates found in the templateDir
// loaded.
func NewDefaultView(templatesDir string, watch bool) (*DefaultView, error) {
	config := &hot.Config{
		Watch:          watch,
		BaseName:       "hero",
		Dir:            templatesDir,
		Funcs:          viewFuncs,
		FilesExtension: []string{".tpl", ".html", ".tmpl"},
	}
	tpl, err := hot.New(config)
	if err != nil {
		return nil, err
	}
	return &DefaultView{tpl: tpl}, nil

}

// Render renders template tplName passing data as context. The result is writen to out.
func (v *DefaultView) Render(out io.Writer, tplName string, data interface{}) error {
	return v.tpl.Execute(out, tplName, data)
}

// Helper for rendering markdown.
func toMarkdown(src string) template.HTML {
	return template.HTML(mark.Render(src))
}
