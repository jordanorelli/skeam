package main

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"fmt"
	"github.com/jordanorelli/skeam/am"
	"html/template"
	"net/http"
	"path/filepath"
)

var assets = am.New("github.com/jordanorelli/skeam", "/static")

func templatePath(relpath string) string {
	return filepath.Join("templates", relpath)
}

func getTemplate(relpath string) (*template.Template, error) {
	p := templatePath(relpath)
	b, err := assets.ReadFile(p)
	if err != nil {
		return nil, errors.New("skeam: unable to read template " + relpath)
	}

	t := template.New(relpath).Funcs(template.FuncMap{
		"js": func(relpath string) (template.HTML, error) {
			jspath := assets.URLPath("js", relpath)
			return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, jspath)), nil
		},
		"css": func(relpath string) (template.HTML, error) {
			csspath := assets.URLPath("css", relpath)
			return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s" />`, csspath)), nil
		},
	})
	t, err = t.Parse(string(b))
	if err != nil {
		return nil, errors.New("skeam: unable to parse template " + relpath + ": " + err.Error())
	}
	return t, err
}

type templateHandler string

func (t templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tpl, err := getTemplate(string(t))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tpl.Execute(w, nil); err != nil {
		fmt.Println(err.Error())
	}
}

func wsHandler(ws *websocket.Conn) {
	manager.Add(ws)
	defer manager.Remove(ws)

	i := newInterpreter(ws, ws, ws)
	i.run(universe)
}

func runHTTPServer() {
	http.Handle("/", templateHandler("home.html"))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(assets.AbsPath("static")))))
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.ListenAndServe(*httpAddr, nil)
}
