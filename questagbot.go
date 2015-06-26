package questagbot

import (
  "appengine"
  "net/http"
  "github.com/codegangsta/martini"
  // hexapic "github.com/blan4/hexapic/core"
)

func AppEngine(c martini.Context, r *http.Request) {
  c.Map(appengine.NewContext(r))
}

func init() {
  m := martini.Classic()
  m.Use(AppEngine)
  m.Get("/", func(c appengine.Context) string {
    return "Hello world"
  })
  http.Handle("/", m)
}