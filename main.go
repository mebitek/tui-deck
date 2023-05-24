package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"tui-deck/deck_http"
	"tui-deck/deck_structs"
	"tui-deck/utils"
)

var app = tview.NewApplication()
var pages = tview.NewPages()
var fullFlex = tview.NewFlex()
var mainFlex = tview.NewFlex()
var footerBar = tview.NewTextView()
var stacks []deck_structs.Stack
var cardsMap = make(map[int]deck_structs.Card)
var detailText = tview.NewTextView()
var detailEditText = tview.NewTextArea()
var primitives = make(map[tview.Primitive]int)
var primitivesIndexMap = make(map[int]tview.Primitive)
var editableObj = deck_structs.Card{}
var currentBoard deck_structs.Board

func main() {

	configFile, err := utils.InitConfingDirectory()
	if err != nil {
		footerBar.SetText(err.Error())
	}

	configuration, err := utils.GetConfiguration(configFile)
	if err != nil {
		footerBar.SetText(err.Error())
	}

	//TODO add default board parameter

	boards, err := deck_http.GetBoards(configuration)
	if err != nil {
		footerBar.SetText(err.Error())
	}

	if len(boards) > 0 {
		currentBoard = boards[0]
	} else {
		footerBar.SetText("No boards found")
	}

	stacks, err = deck_http.GetStacks(currentBoard.Id, configuration)
	if err != nil {
		footerBar.SetText(err.Error())
	}

	go buildStacks()

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
			stacks, err = deck_http.GetStacks(currentBoard.Id, configuration)
			if err != nil {
				footerBar.SetText(err.Error())
			}

			go buildStacks()
		}
		return event
	})

	mainFlex.SetTitle(" TUI TODO ")
	mainFlex.SetDirection(tview.FlexColumn)
	mainFlex.SetBorder(true)
	mainFlex.SetBorderColor(tcell.Color133)

	footerBar.SetBorder(true)
	footerBar.SetTitle(" Info ")
	footerBar.SetBorderColor(tcell.Color133)

	fullFlex.SetDirection(tview.FlexRow)
	fullFlex.AddItem(mainFlex, 0, 10, true)
	fullFlex.AddItem(footerBar, 0, 1, false)

	detailText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			go buildFullFlex(mainFlex)
		} else if event.Rune() == 101 {
			//editableObj = todoByUidMaps[editableObj.Index]
			detailEditText.SetTitle(" " + detailText.GetTitle() + " - EDIT ")
			detailEditText.SetText(formatDescription(editableObj.Description), true)
			go buildFullFlex(detailEditText)
		}
		return event
	})

	detailEditText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			detailText.Clear()
			detailText.SetTitle(" " + editableObj.Title + " ")
			detailText.SetText(formatDescription(editableObj.Description))

			go buildFullFlex(detailText)

		} else if event.Key() == tcell.KeyF2 {
			editableObj.Description = detailEditText.GetText()

			jsonBody := strings.ReplaceAll(`{"description": "`+editableObj.Description+`", "title": "`+editableObj.Title+`", "type": "plain", "owner":"`+configuration.User+`"}`, "\n", `\n`)

			editableObj, err = deck_http.UpdateCard(currentBoard.Id, editableObj.StackId, editableObj.Id, jsonBody, configuration)
			if err != nil {
				footerBar.SetText(err.Error())
				return event
			}

			cardsMap[editableObj.Id] = editableObj

			detailText.SetText(formatDescription(editableObj.Description))
			go buildFullFlex(detailText)

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

func buildStacks() {
	for index, s := range stacks {

		todoList := tview.NewList()
		todoList.SetTitle(" " + s.Title + " ")
		todoList.SetBorder(true)
		//todoList.SetSecondaryTextColor(tcell.Color133)

		todoList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTAB {
				return nil
			}
			return event
		})

		for _, card := range s.Cards {
			var labels = ""
			for i, label := range card.Labels {
				labels = labels + "[#" + label.Color + "]" + label.Title + "[white]"
				if i != len(card.Labels)-1 {
					labels = labels + ", "
				}
			}
			cardsMap[card.Id] = card

			todoList.AddItem(strconv.Itoa(card.Id)+" - "+card.Title, labels, rune(0), nil)
		}

		todoList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {

			split := strings.Split(name, "-")
			cardId, _ := strconv.Atoi(strings.TrimSpace(split[0]))

			detailText.SetTitle(" " + cardsMap[cardId].Title + " ")
			detailText.SetText(formatDescription(cardsMap[cardId].Description))
			editableObj = cardsMap[cardId]
			buildFullFlex(detailText)
		})

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

func formatDescription(description string) string {
	return strings.ReplaceAll(description, `\n`, "\n")
}

func buildFullFlex(primitive tview.Primitive) {
	fullFlex.Clear()
	fullFlex.AddItem(primitive, 0, 10, true)
	fullFlex.AddItem(footerBar, 0, 1, false)
	app.SetFocus(primitive)
}
