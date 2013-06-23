package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
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

type templateHandler struct {
	path    string
	context interface{}
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tpl, err := getTemplate(t.path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tpl.Execute(w, t.context); err != nil {
		fmt.Println(err.Error())
	}
}

type wsMessage struct {
	IsError bool   `json:"is_error"`
	Message string `json:"message"`
}

type wsWriter struct {
	conn *websocket.Conn
}

func (w wsWriter) Write(b []byte) (int, error) {
	out, err := json.Marshal(wsMessage{false, string(b)})
	if err != nil {
		return 0, err
	}
	return w.conn.Write(out)
}

type wsErrorWriter wsWriter

func (w wsErrorWriter) Write(b []byte) (int, error) {
	out, err := json.Marshal(wsMessage{true, string(b)})
	if err != nil {
		return 0, err
	}
	return w.conn.Write(out)
}

func wsHandler(ws *websocket.Conn) {
	manager.Add(ws)
	defer manager.Remove(ws)

	i := newInterpreter(ws, wsWriter{ws}, wsErrorWriter{ws})
	i.run(universe)
}

func runHTTPServer() {
	http.Handle("/", &templateHandler{"home.html", map[string]interface{}{"ws_path": template.JS("/ws")}})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(assets.AbsPath("static")))))
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.ListenAndServe(*httpAddr, nil)
}
