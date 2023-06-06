package deck_ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tui-deck/deck_help"
	"tui-deck/utils"
)

const VERSION = "v0.5.3"

var FullFlex = tview.NewFlex()
var MainFlex = tview.NewFlex()
var FooterBar = tview.NewTextView()

var Primitives = make(map[tview.Primitive]int)
var PrimitivesIndexMap = make(map[int]tview.Primitive)
var app *tview.Application
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration) {
	app = application
	configuration = conf

	MainFlex.SetDirection(tview.FlexColumn)
	MainFlex.SetBorder(true)
	MainFlex.SetBorderColor(utils.GetColor(configuration.Color))

	FooterBar.SetBorder(true)
	FooterBar.SetTitle(" Info ")
	FooterBar.SetBorderColor(utils.GetColor(configuration.Color))
	FooterBar.SetDynamicColors(true)
	FooterBar.SetText("Press [yellow]?[white] for help, [yellow]q[white] to exit")

	FullFlex.SetDirection(tview.FlexRow)
	FullFlex.AddItem(MainFlex, 0, 10, true)
	FullFlex.AddItem(FooterBar, 0, 1, false)
}

func BuildFullFlex(primitive tview.Primitive) {
	FullFlex.Clear()
	FullFlex.AddItem(primitive, 0, 10, true)
	FullFlex.AddItem(FooterBar, 0, 1, false)
	if primitive != MainFlex {
		FooterBar.SetText("Press [yellow]?[white] for help, [yellow]ESC[white] to go back")

	} else {
		FooterBar.SetText("Press [yellow]?[white] for help, [yellow]q[white] to exit")
	}
	app.SetFocus(primitive)
}

func BuildHelp(primitive tview.Primitive, helpView *tview.TextView) {
	help := tview.NewFrame(helpView)
	help.SetBorder(true)
	help.SetBorderColor(utils.GetColor(configuration.Color))
	help.SetTitle(helpView.GetTitle())
	FooterBar.SetTitle(VERSION)
	BuildFullFlex(help)

	help.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			BuildFullFlex(primitive)
			FooterBar.SetTitle(" Info ")
			return nil
		} else if event.Key() == tcell.KeyEnter {
			switch {
			case help.GetPrimitive() == deck_help.HelpMain:
				help.SetTitle(deck_help.HelpView.GetTitle())
				help.SetPrimitive(deck_help.HelpView)
				return nil
			case help.GetPrimitive() == deck_help.HelpView:
				help.SetTitle(deck_help.HelpEdit.GetTitle())
				help.SetPrimitive(deck_help.HelpEdit)
				return nil
			case help.GetPrimitive() == deck_help.HelpEdit:
				help.SetTitle(deck_help.HelpLabels.GetTitle())
				help.SetPrimitive(deck_help.HelpLabels)
				return nil
			case help.GetPrimitive() == deck_help.HelpLabels:
				help.SetTitle(deck_help.HelpComments.GetTitle())
				help.SetPrimitive(deck_help.HelpComments)
				return nil
			case help.GetPrimitive() == deck_help.HelpComments:
				help.SetTitle(deck_help.HelpBoards.GetTitle())
				help.SetPrimitive(deck_help.HelpBoards)
				return nil
			case help.GetPrimitive() == deck_help.HelpBoards:
				help.SetTitle(deck_help.HelpMain.GetTitle())
				help.SetPrimitive(deck_help.HelpMain)
				return nil
			}
		}
		return event
	})
}

func GetNextFocus(index int) tview.Primitive {
	if index == len(PrimitivesIndexMap) {
		index = 0
	}
	return PrimitivesIndexMap[index]
}
