package hero

import (
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/a8m/mark"
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
	tpl *template.Template
}

// NewDefaultView returns a a DefaultView instance with all templates found in the templateDir
// loaded.
func NewDefaultView(templatesDir string) (*DefaultView, error) {
	tpl := &DefaultView{
		tpl: template.New("hero").Funcs(viewFuncs),
	}
	if err := tpl.load(templatesDir); err != nil {
		return nil, err
	}
	return tpl, nil

}

// load parses templates from templateDir
func (v *DefaultView) load(templatesDir string) error {
	stat, serr := os.Stat(templatesDir)
	if serr != nil {
		return serr
	}
	if !stat.IsDir() {
		return errors.New("hero: invalid template directory")
	}
	return filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(path, templatesDir)
		name = filepath.ToSlash(name)
		name = strings.TrimPrefix(name, "/")
		t := v.tpl.New(name)
		_, err = t.Parse(string(data))
		if err != nil {
			return err
		}
		return nil
	})
}

// Render renders template tplName passing data as context. The result is writen to out.
func (v *DefaultView) Render(out io.Writer, tplName string, data interface{}) error {
	return v.tpl.ExecuteTemplate(out, tplName, data)
}

// Helper for rendering markdown.
func toMarkdown(src string) template.HTML {
	return template.HTML(mark.Render(src))
}
