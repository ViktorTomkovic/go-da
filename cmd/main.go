package main

import (
	"html/template"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Teammate struct {
	Name           string
	Done           string
	WillDo         string
	Blockers       string
	GeneralRemarks string
}

func (t Teammate) DashedId() string {
	return strings.Replace(t.Name, " ", "-", -1)
}

func (t Teammate) IsSelected(selectedName string) bool {
	return t.Name == selectedName
}

func (t Teammate) SelectedClass(selectedName string) string {
	if t.IsSelected(selectedName) {
		return " selected"
	}
	return ""
}

type Teammates = []Teammate

func newTeammates() []Teammate {
	return []Teammate{
		{
			Name:           "Viktor",
			Done:           "asdf",
			WillDo:         "fdas",
			Blockers:       "nothing",
			GeneralRemarks: "",
		},
		{
			Name:           "Viktor Tomkovič",
			Done:           "asdšľč čšľ ľľľ§ňäf",
			WillDo:         "žšť,,-.z jfdas",
			Blockers:       "nothing\" <div>asdlf</div> \"",
			GeneralRemarks: "rrrrr",
		},
	}
}

type Page struct {
	Teammates []Teammate
	SelectedName string
	ActivatedBy string
}

func newPage() Page {
	return Page{
		Teammates: newTeammates(),
		SelectedName: "",
		ActivatedBy: "",
	}
}

func main() {
	server := echo.New()
	server.Use(middleware.Logger())
	page := newPage()
	server.File("/main.css", "css/main.css")
	server.File("/favicon.png", "favicon.png")
	server.Renderer = newTemplate()
	server.GET("/", func(c echo.Context) error {
		selectedName := c.QueryParam("selectedName")
		activatedBy := c.QueryParam("activatedBy")
		page.SelectedName = selectedName
		page.ActivatedBy = activatedBy
		return c.Render(200, "index", page)
	})
	err := server.Start(":12121")
	server.Logger.Fatal(err)
}
