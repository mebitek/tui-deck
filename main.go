package main

import (
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sort"
	"strconv"
	"strings"
	"tui-deck/utils"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var fullFlex = tview.NewFlex()
var mainFlex = tview.NewFlex()
var footerBar = tview.NewTextView()
var stacks = make([]VTodoObject, 0)
var todoMaps = make(map[string][]VTodoObject)
var todoByUidMaps = make(map[int]VTodoObject)
var detailText = tview.NewTextView()
var detailEditText = tview.NewTextArea()
var primitives = make(map[tview.Primitive]int)
var primitivesIndexMap = make(map[int]tview.Primitive)
var editableObj = VTodoObject{}
var calendarClient = caldav.Client{}

type VTodoObject struct {
	Path        string
	Index       int
	DtStamp     string
	Uid         string
	RelatedTo   string
	Status      string
	Categories  string
	Summary     string
	Description string
}

func main() {

	configFile, err := utils.InitConfingDirectory()
	if err != nil {
		footerBar.SetText(err.Error())
	}

	configuration, err := utils.GetConfiguration(configFile)
	if err != nil {
		footerBar.SetText(err.Error())
	}

	calendarClient = loadCalendars(configuration)

	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 {
			// q
			app.Stop()
		} else if event.Key() == tcell.KeyTab {
			// tab
			primitive := app.GetFocus()
			list := primitive.(*tview.List)
			list.SetTitleColor(tcell.ColorWhite)

			actualPrimitiveIndex := primitives[primitive]
			app.SetFocus(getNextFocus(actualPrimitiveIndex + 1))

		} else if event.Rune() == 114 {
			// r
			calendarClient = loadCalendars(configuration)
		}
		return event
	})

	mainFlex.SetTitle("TUI TODO")
	mainFlex.SetDirection(tview.FlexColumn)
	mainFlex.SetBorder(true)
	mainFlex.SetBorderColor(tcell.Color133)

	footerBar.SetBorder(true)
	footerBar.SetTitle("Info")
	footerBar.SetBorderColor(tcell.Color133)

	fullFlex.SetDirection(tview.FlexRow)
	fullFlex.AddItem(mainFlex, 0, 10, true)
	fullFlex.AddItem(footerBar, 0, 1, false)

	detailText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			buildFullFlex(mainFlex)
		} else if event.Rune() == 101 {
			editableObj = todoByUidMaps[editableObj.Index]
			detailEditText.SetTitle(detailText.GetTitle() + " - EDIT")
			detailEditText.SetText(formatDescription(editableObj.Description), true)
			buildFullFlex(detailEditText)
		}
		return event
	})

	detailEditText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			detailText.Clear()
			detailText.SetTitle(editableObj.Summary)
			detailText.SetText(formatDescription(editableObj.Description))

			buildFullFlex(detailText)

		} else if event.Key() == tcell.KeyF2 {
			editableObj.Description = detailEditText.GetText()
			_, err := calendarClient.PutCalendarObject(editableObj.Path, buildICal())
			if err != nil {
				footerBar.SetText(err.Error())
			}

		}
		return event
	})
	detailText.SetBorder(true)
	detailText.SetBorderColor(tcell.Color133)

	detailEditText.SetBorder(true)
	detailEditText.SetBorderColor(tcell.Color133)

	pages.AddPage("Main", fullFlex, true, true)

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

func loadCalendars(configuration utils.Configuration) caldav.Client {
	mainFlex.Clear()
	stacks = make([]VTodoObject, 0)
	todoMaps = make(map[string][]VTodoObject)
	auth := webdav.HTTPClientWithBasicAuth(nil, configuration.User, configuration.Password)

	cal, err := caldav.NewClient(auth, configuration.Url)
	if err != nil {
		footerBar.SetText(err.Error())

	}
	calendars, err := cal.FindCalendars("")
	if calendars == nil {
		footerBar.SetText("No calendars found. Please check your configuration")
	}

	for _, c := range calendars {

		if contains(c.SupportedComponentSet, "VTODO") {
			query := &caldav.CalendarQuery{
				CompFilter: caldav.CompFilter{
					Name: "VCALENDAR",
					Comps: []caldav.CompFilter{
						{
							Name: "VTODO",
						},
					},
				},
			}

			var calendarObjects, err = cal.QueryCalendar(c.Path, query)
			if err != nil {
				footerBar.SetText(err.Error())
			}

			for _, c := range calendarObjects {
				props := c.Data.Children[0].Props
				uid := props.Get("UID").Value
				split := strings.Split(uid, "-")
				index, _ := strconv.Atoi(split[len(split)-1])
				obj := VTodoObject{
					Path:    c.Path,
					Index:   index,
					DtStamp: props.Get("DTSTAMP").Value,
					Uid:     uid,
					Summary: props.Get("SUMMARY").Value,
				}
				if props.Get("RELATED-TO") == nil {
					stacks = append(stacks, obj)
				} else {
					obj.Description = props.Get("DESCRIPTION").Value
					if props.Get("CATEGORIES") != nil {
						obj.Categories = props.Get("CATEGORIES").Value
					}
					obj.Status = props.Get("STATUS").Value
					obj.RelatedTo = props.Get("RELATED-TO").Value
					_, exists := todoMaps[obj.RelatedTo]
					if !exists {
						todoMaps[obj.RelatedTo] = make([]VTodoObject, 0)
					}
					todoMaps[obj.RelatedTo] = append(todoMaps[obj.RelatedTo], obj)
				}
				todoByUidMaps[obj.Index] = obj
			}
		}
	}
	buildStacks()
	return *cal
}

func buildStacks() {
	for index, s := range stacks {

		list := todoMaps[s.Uid]

		sort.Slice(list, func(i, j int) bool {
			return list[i].Index > (list[j].Index)
		})

		todoList := tview.NewList()
		todoList.SetTitle(s.Summary)
		todoList.SetBorder(true)
		todoList.SetSecondaryTextColor(tcell.Color133)

		todoList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			detailText.SetTitle(list[index].Summary)
			detailText.SetText(formatDescription(list[index].Description))
			editableObj = list[index]
			buildFullFlex(detailText)
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

		todoList.SetFocusFunc(func() {
			todoList.SetTitleColor(tcell.Color133)
		})

		primitives[todoList] = index
		primitivesIndexMap[index] = todoList

		mainFlex.AddItem(todoList, 0, 1, true)
		primitive := mainFlex.GetItem(0)
		app.SetFocus(primitive)
	}
}

func buildICal() *ical.Calendar {
	return &ical.Calendar{Component: &ical.Component{
		Name: "VCALENDAR",
		Props: ical.Props{
			"PRODID": []ical.Prop{{
				Name:   "PRODID",
				Params: ical.Params{},
				Value:  "-//Sabre//Sabre VObject 4.4.2//EN",
			}},
			"VERSION": []ical.Prop{{
				Name:   "VERSION",
				Params: ical.Params{},
				Value:  "2.0",
			}},
			"CALSCALE": []ical.Prop{{
				Name:   "CALSCALE",
				Params: ical.Params{},
				Value:  "GREGORIAN",
			}},
		},
		Children: []*ical.Component{
			{
				Name: "VTODO",
				Props: ical.Props{
					"DTSTAMP": []ical.Prop{{
						Name:   "DTSTAMP",
						Params: ical.Params{},
						Value:  editableObj.DtStamp,
					}},
					"UID": []ical.Prop{{
						Name:   "UID",
						Params: ical.Params{},
						Value:  editableObj.Uid,
					}},
					"RELATED-TO": []ical.Prop{{
						Name:   "RELATED-TO",
						Params: ical.Params{},
						Value:  editableObj.RelatedTo,
					}},
					"STATUS": []ical.Prop{{
						Name:   "STATUS",
						Params: ical.Params{},
						Value:  editableObj.Status,
					}},
					"CATEGORIES": []ical.Prop{{
						Name:   "CATEGORIES",
						Params: ical.Params{},
						Value:  editableObj.Categories,
					}},
					"SUMMARY": []ical.Prop{{
						Name:   "SUMMARY",
						Params: ical.Params{},
						Value:  editableObj.Summary,
					}},
					"DESCRIPTION": []ical.Prop{{
						Name:   "DESCRIPTION",
						Params: ical.Params{},
						Value:  editableObj.Description,
					}},
				},
			},
		},
	}}
}

func formatDescription(description string) string {
	return strings.ReplaceAll(description, `\n`, "\n")
}

func buildFullFlex(primitive tview.Primitive) {
	fullFlex.Clear()
	fullFlex.AddItem(primitive, 0, 10, true)
	fullFlex.AddItem(footerBar, 0, 1, false)
	app.SetFocus(primitive)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
