package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
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

func newDefaultTeammates() []Teammate {
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

func newTeammate(name string) Teammate {
	return Teammate{
		Name: name,
		Done: "",
		WillDo: "",
		Blockers: "",
		GeneralRemarks: "",
	}
}

func newNamedTeammates(names []string) []Teammate {
	result := []Teammate{}
	for _, name := range names {
		result = append(result, newTeammate(name))
	}
	return result
}

type Page struct {
	Teammates    []Teammate
	SelectedName string
	ActivatedBy  string
}

func newPage() Page {
	return Page{
		Teammates:    newDefaultTeammates(),
		SelectedName: "",
		ActivatedBy:  "",
	}
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

type Config struct {
	Port  int
	Names []string
}

func CreateDeafultConfig() Config {
	return Config{
		Port:  12121,
		Names: []string{"John Blow", "Johan Brahms", "Ján Hraško"},
	}
}

func initConfig() Config {
	result := CreateDeafultConfig()
	if _, err := os.Stat(".da.json"); err == nil {
		// exists
		fileContent, err := os.ReadFile(".da.json")
		if err == nil {
			var config Config
			err := json.Unmarshal(fileContent, &config)
			if err != nil {
				fmt.Println("Error during unmarshalling json from config file .da.json. Using default configuration.")
			} else {
				result = config
			}
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// path does *not* exist
		fmt.Println("File .da.json does not exists. Using default configuration.")
	} else {
		// File cannot be reached. See err for details.
		fmt.Println("File .da.json cannot be reached. Using default configuration.")
	}
	return result
}

const TeammatesFilename = "daily.json"

func writeTeammates(teammates []Teammate) {
	content, err := json.MarshalIndent(teammates, "", "\t")
	if err != nil {
		fmt.Println("CANNOT MARSHAL daily.json!")
		fmt.Println(err)
	} else {
		err = os.WriteFile(TeammatesFilename, content, 0644)
		if err != nil {
			fmt.Println("CANNOT WRITE daily.json!")
			fmt.Println(err)
		}
	}
}

func readTeammates() ([]Teammate, error) {
	content, err := os.ReadFile(TeammatesFilename)
	if err != nil {
		return []Teammate{}, err
	}
	var teammates []Teammate
	err = json.Unmarshal(content, &teammates)
	if err != nil {
		return []Teammate{}, err
	}
	return teammates, nil
}

func main() {
	fmt.Println("Daily Automation. Welcome and have a nice productive day.")
	config := initConfig()
	teammatesInit, err := readTeammates()
	page := newPage()
	if err == nil {
		page.Teammates = teammatesInit
	} else {
		page.Teammates = newNamedTeammates(config.Names)
	}
	fmt.Println(config.Names)
	fmt.Println(teammatesInit)
	fmt.Println(page.Teammates)
	server := echo.New()
	server.Use(middleware.Logger())
	server.File("/main.css", "css/main.css")
	server.File("/favicon.png", "favicon.png")
	server.Renderer = newTemplate()
	server.GET("/", func(c echo.Context) error {
		selectedName := c.QueryParam("selectedName")
		activatedBy := c.QueryParam("activatedBy")
		page.SelectedName = selectedName
		page.ActivatedBy = activatedBy
		teammatesFromFile, err := readTeammates()
		if err != nil {
			fmt.Println("Could not read teammates from daily.json. Using default teammates.")
			fmt.Println(err)
		} else {
			page.Teammates = teammatesFromFile
		}
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
				} else {
					page.Teammates = append(page.Teammates, modifiedTeammate)
				}
			}
			// Write down changes to harddrive
			cloneTeammates := page.Teammates
			writeTeammates(cloneTeammates)
		}
		return c.Render(200, "index", page)
	})
	err = server.Start(fmt.Sprintf(":%d", config.Port))
	server.Logger.Fatal(err)
	fmt.Println("Daily Automation. Bye and have a nice productive day.")
}
