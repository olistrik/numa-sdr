package filesystem

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileSystem struct {
	FileSystem http.FileSystem
}

// Open passes `Open` to the upstream implementation and return an [fs.File].
func (o FileSystem) Open(name string) (fs.File, error) {
	f, err := o.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	return fs.File(f), nil
}

// LoadHTMLFS loads an http.FileSystem and a slice of patterns
// and associates the result with HTML renderer.
func LoadHTMLFS(engine *gin.Engine, fs http.FileSystem, patterns ...string) {
	// if gin.IsDebugging() {
	// 	engine.HTMLRender = render.HTMLDebug{FileSystem: fs, Patterns: patterns, FuncMap: engine.FuncMap, Delims: engine.delims}
	// 	return
	// }

	templ := template.Must(template.New("").Delims("{{", "}}").Funcs(engine.FuncMap).ParseFS(
		FileSystem{FileSystem: fs}, patterns...))
	engine.SetHTMLTemplate(templ)
}
