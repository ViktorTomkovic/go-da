package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"slices"
	"strconv"
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
	Teammates    []Teammate
	SelectedName string
	ActivatedBy  string
}

func newPage() Page {
	return Page{
		Teammates:    newTeammates(),
		SelectedName: "",
		ActivatedBy:  "",
	}
}

func main() {
	fmt.Println("Daily Automation. Welcome and have a nice productive day.")
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
		numberSelected, _ := strconv.Atoi(c.QueryParam("numberSelected"))
		if numberSelected > 0 {
			// We moved from editing one teammate to another. Update recently edited one.
			modifiedTeammate, err := createTeammate(
				c.QueryParam("name"),
				c.QueryParam("done"),
				c.QueryParam("willDo"),
				c.QueryParam("blockers"),
				c.QueryParam("generalRemarks"))
			if err == nil {
				i := slices.IndexFunc(page.Teammates, func(t Teammate) bool {
					return t.Name == modifiedTeammate.Name
				})
				if i >= 0 {
					page.Teammates[i] = modifiedTeammate
				}
			}
		}
		return c.Render(200, "index", page)
	})
	err := server.Start(":12121")
	server.Logger.Fatal(err)
	fmt.Println("Daily Automation. Bye and have a nice productive day.")
}

func createTeammate(name string, done string, willDo string, blockers string, generalRemarks string) (Teammate, error) {
	if name == "" {
		return Teammate{}, errors.New("Empty name")
	}
	return Teammate{
		Name:           name,
		Done:           done,
		WillDo:         willDo,
		Blockers:       blockers,
		GeneralRemarks: generalRemarks,
	}, nil
}
