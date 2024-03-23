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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
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
		Name:           name,
		Done:           "",
		WillDo:         "",
		Blockers:       "",
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

func writeTeammates(path string, teammates []Teammate) error {
	content, err := json.MarshalIndent(teammates, "", "\t")
	if err != nil {
		fmt.Println("CANNOT MARSHAL TEAMMATES JSON!")
		return err
	} else {
		err = os.WriteFile(path, content, 0644)
		if err != nil {
			fmt.Printf("CANNOT WRITE '%s'!\n", path)
			return err
		}
	}
	return nil
}

func readTeammates(path string) ([]Teammate, error) {
	content, err := os.ReadFile(path)
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

func resetDailyEntries() {
	time := time.Now()
	ts := time.Format("20060102150405")
	resetFilename := fmt.Sprintf("daily.%s.json", ts)
	err := os.Mkdir("oldDailies", os.ModeDir)
	if err != nil {
		fmt.Println(err)
		return
	}
	resetFilepath := fmt.Sprintf("oldDailies/%s", resetFilename)
	teammates, err := readTeammates(TeammatesFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = writeTeammates(resetFilepath, teammates)
	if err != nil {
		fmt.Println(err)
		return
	}
	resetTeammates := []Teammate{}
	for _, teammate := range teammates {
		resetTeammates = append(resetTeammates, newTeammate(teammate.Name))
	}
	err = writeTeammates(TeammatesFilename, resetTeammates)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	fmt.Println("Daily Automation. Welcome and have a nice productive day.")
	config := initConfig()
	teammatesInit, err := readTeammates(TeammatesFilename)
	page := newPage()
	if err == nil {
		page.Teammates = teammatesInit
	} else {
		page.Teammates = newNamedTeammates(config.Names)
	}
	// What teammates are used
	fmt.Println(page.Teammates)
	// Set up cron
	c := cron.New()
	/*
	_, err = c.AddFunc("* * * * *", resetDailyEntries)
	if err != nil {
		fmt.Println(err)
	}
	*/
	_, err = c.AddFunc("@daily", resetDailyEntries)
	if err != nil {
		fmt.Println(err)
	}
	c.Start()
	// Set up server
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
		teammatesFromFile, err := readTeammates(TeammatesFilename)
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
			writeTeammates(TeammatesFilename, cloneTeammates)
		}
		return c.Render(200, "index", page)
	})
	err = server.Start(fmt.Sprintf(":%d", config.Port))
	server.Logger.Fatal(err)
	c.Stop()
	fmt.Println("Daily Automation. Bye and have a nice productive day.")
}
