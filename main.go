package main

import (
	"encoding/json"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"tui-deck/utils"
)

var app = tview.NewApplication()
var pages = tview.NewPages()

var mainFlex = tview.NewFlex()
var stacks = make([]VTodoObect, 0)

var todoMaps = make(map[string][]VTodoObect)

var detailText = tview.NewTextView()

var primitives = make(map[tview.Primitive]int)
var primitivesIndexMap = make(map[int]tview.Primitive)

type Configuration struct {
	User     string `json:"username"`
	Password string `json:"password"`
	Url      string `json:"url"`
}

type VTodoObect struct {
	Index       int
	DtStamp     time.Time
	Uid         string
	RelatedTo   string
	Status      string
	Categories  string
	Summary     string
	Description string
}

func main() {

	configFile := utils.InitConfingDirectory()

	file, _ := os.Open(configFile)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	auth := webdav.HTTPClientWithBasicAuth(nil, configuration.User, configuration.Password)

	cal, err := caldav.NewClient(auth, configuration.Url)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	calendars, err := cal.FindCalendars("")

	for _, c := range calendars {
		if c.SupportedComponentSet[0] == "VTODO" {
			query := &caldav.CalendarQuery{
				CompFilter: caldav.CompFilter{
					Name: "VCALENDAR",
					Comps: []caldav.CompFilter{{
						Name: "VTODO",
					},
					},
				},
			}

			calendarObjecrs, err := cal.QueryCalendar(c.Path, query)
			if err != nil {
				log.Fatal("Error when opening file: ", err)
			}

			for _, c := range calendarObjecrs {
				props := c.Data.Children[0].Props
				uid := props.Get("UID").Value
				split := strings.Split(uid, "-")
				index, _ := strconv.Atoi(split[len(split)-1])
				obj := VTodoObect{
					Index:   index,
					DtStamp: c.ModTime,
					Uid:     uid,
					Summary: props.Get("SUMMARY").Value,
				}
				if props.Get("RELATED-TO") == nil {
					stacks = append(stacks, obj)
				} else {
					obj.Description = props.Get("DESCRIPTION").Value
					obj.Categories = props.Get("CATEGORIES").Value
					obj.Status = props.Get("STATUS").Value
					obj.RelatedTo = props.Get("RELATED-TO").Value
					_, exists := todoMaps[obj.RelatedTo]
					if !exists {
						todoMaps[obj.RelatedTo] = make([]VTodoObect, 0)
					}
					todoMaps[obj.RelatedTo] = append(todoMaps[obj.RelatedTo], obj)
				}
			}
		}
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 {
			// q
			app.Stop()
		} else if event.Key() == tcell.KeyTab {
			primitive := app.GetFocus()
			actualPrimitiveIndex := primitives[primitive]
			app.SetFocus(getNextFocus(actualPrimitiveIndex + 1))

		}
		return event
	})

	mainFlex.SetTitle("TUI TODO")
	mainFlex.SetDirection(tview.FlexColumn)
	mainFlex.SetBorder(true)
	mainFlex.SetBorderColor(tcell.Color133)
	for index, s := range stacks {
		flex := tview.NewFlex()
		flex.SetTitle(s.Summary)
		flex.SetBorder(true)

		uid := s.Uid

		list := todoMaps[uid]

		sort.Slice(list, func(i, j int) bool {
			return list[i].Index > (list[j].Index)
		})

		todoList := tview.NewList()
		todoList.SetSecondaryTextColor(tcell.Color133)

		todoList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			detailText.SetTitle(list[index].Summary)
			detailText.SetText(strings.ReplaceAll(list[index].Description, `\n`, "\n"))
			pages.SwitchToPage("TodoDetail")
		})

		todoList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTAB {
				return nil
			}
			return event
		})
		for _, l := range list {
			todoList.AddItem(strconv.Itoa(l.Index)+" - "+l.Summary, l.Categories, rune(0), nil)
		}

		flex.AddItem(todoList, 0, 1, true)
		primitives[todoList] = index
		primitivesIndexMap[index] = todoList
		mainFlex.AddItem(flex, 0, 1, true)
	}

	detailText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("Main")
		}
		return event
	})
	detailText.SetBorder(true)
	detailText.SetBorderColor(tcell.Color133)

	pages.AddPage("Main", mainFlex, true, true)
	pages.AddPage("TodoDetail", detailText, true, false)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

func getNextFocus(index int) tview.Primitive {

	if index == len(primitivesIndexMap) {
		index = 0
	}
	return primitivesIndexMap[index]
}
