package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"tui-deck/deck_help"
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
var boards []deck_structs.Board
var cardsMap = make(map[int]deck_structs.Card)
var detailText = tview.NewTextView()
var detailEditText = tview.NewTextArea()
var primitives = make(map[tview.Primitive]int)
var primitivesIndexMap = make(map[int]tview.Primitive)
var editableCard = deck_structs.Card{}
var currentBoard deck_structs.Board
var boardList = tview.NewList()

var editTagsFlex = tview.NewFlex()

var configuration utils.Configuration

func main() {
	deck_help.InitHelp()
	configFile, err := utils.InitConfingDirectory()
	if err != nil {
		footerBar.SetText(err.Error())
	}

	configuration, err = utils.GetConfiguration(configFile)
	if err != nil {
		footerBar.SetText(err.Error())
	}

	//TODO add default board parameter?
	fmt.Print("Getting boards...\n")
	boards, err = deck_http.GetBoards(configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error getting boards: %s", err.Error()))
	}

	if len(boards) > 0 {
		fmt.Print("Getting current board detail...\n")
		currentBoard, err = deck_http.GetBoardDetail(boards[0].Id, configuration)
		if err != nil {
			footerBar.SetText(fmt.Sprintf("Error getting board detail: %s", err.Error()))

		}
		go buildSwitchBoard(configuration)
	} else {
		footerBar.SetText("No boards found")
	}

	fmt.Print("Getting stacks...\n")
	stacks, err = deck_http.GetStacks(currentBoard.Id, configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
	}

	go buildStacks()

	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 {
			// q -> quit app
			app.Stop()
		} else if event.Key() == tcell.KeyTab {
			// tab -> switch focus between stacks
			primitive := app.GetFocus()
			list := primitive.(*tview.List)
			list.SetTitleColor(tcell.ColorWhite)

			actualPrimitiveIndex := primitives[primitive]
			app.SetFocus(getNextFocus(actualPrimitiveIndex + 1))

		} else if event.Rune() == 114 {
			// r -> reload stacks
			stacks, err = deck_http.GetStacks(currentBoard.Id, configuration)
			if err != nil {
				footerBar.SetText(fmt.Sprintf("Error reloading stacks: %s", err.Error()))
			}
			go buildStacks()
		} else if event.Rune() == 115 {
			// s -> switch board
			go buildFullFlex(boardList)
		} else if event.Rune() == 63 {
			// ? deck_help menu
			buildHelp(mainFlex, deck_help.HelpMain)
		}

		return event

	})

	mainFlex.SetTitle(fmt.Sprintf(" TUI DECK: [#%s]%s ", currentBoard.Color, currentBoard.Title))
	mainFlex.SetDirection(tview.FlexColumn)
	mainFlex.SetBorder(true)
	mainFlex.SetBorderColor(utils.GetColor(configuration.Color))

	footerBar.SetBorder(true)
	footerBar.SetTitle(" Info ")
	footerBar.SetBorderColor(utils.GetColor(configuration.Color))
	footerBar.SetDynamicColors(true)
	footerBar.SetText("Press [yellow]?[white] for help, [yellow]q[white] to exit")

	fullFlex.SetDirection(tview.FlexRow)
	fullFlex.AddItem(mainFlex, 0, 10, true)
	fullFlex.AddItem(footerBar, 0, 1, false)

	detailText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// ESC -> back to main view
			go buildFullFlex(mainFlex)
		} else if event.Rune() == 101 {
			// e -> edit description
			detailEditText.SetTitle(fmt.Sprintf(" %s- EDIT", detailText.GetTitle()))
			detailEditText.SetText(formatDescription(editableCard.Description), true)
			go buildFullFlex(detailEditText)
		} else if event.Rune() == 116 {
			// t -> tags
			editTagsFlex.Clear()
			actualLabelList := tview.NewList()
			actualLabelList.SetBorder(true)
			actualLabelList.SetTitle(" delete labels ")
			actualLabelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyTab {
					return nil
				}
				return event
			})
			for _, label := range editableCard.Labels {
				actualLabelList.AddItem(fmt.Sprintf("[#%s]%s", label.Color, label.Title), "",
					rune(0), nil)
			}
			actualLabelList.SetSelectedFunc(func(index int, name string, secondName string, rune rune) {
				label := editableCard.Labels[index]
				jsonBody := fmt.Sprintf(`{"labelId": %d}`, label.Id)
				go deleteLabel(jsonBody)
				editableCard.Labels = append(editableCard.Labels[:index], editableCard.Labels[index+1:]...)
				cardsMap[editableCard.Id] = editableCard
				actualLabelList.RemoveItem(index)

				updateStacks()
				buildStacks()
				app.SetFocus(actualLabelList)
			})

			labelList := tview.NewList()
			labelList.SetBorder(true)
			labelList.SetTitle(" add labels")
			labelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyTab {
					return nil
				}
				return event
			})
			for _, label := range currentBoard.Labels {
				labelList.AddItem(fmt.Sprintf("[#%s]%s", label.Color, label.Title), "",
					rune(0), nil)
			}

			labelList.SetSelectedFunc(func(index int, name string, secondName string, rune rune) {
				label := currentBoard.Labels[index]
				jsonBody := fmt.Sprintf(`{"labelId": %d }`, label.Id)
				go assignLabel(jsonBody)
				editableCard.Labels = append(editableCard.Labels, label)
				cardsMap[editableCard.Id] = editableCard
				actualLabelList.AddItem(fmt.Sprintf("[#%s]%s", label.Color, label.Title), "",
					rune, nil)
				updateStacks()
				buildStacks()
				app.SetFocus(labelList)
			})

			editTagsFlex.SetDirection(tview.FlexColumn)
			editTagsFlex.SetBorder(true)
			editTagsFlex.SetBorderColor(utils.GetColor(configuration.Color))
			editTagsFlex.SetTitle(fmt.Sprintf(" %s- EDIT TAGS", detailText.GetTitle()))

			editTagsFlex.AddItem(actualLabelList, 0, 1, true)
			editTagsFlex.AddItem(labelList, 0, 1, true)
			editTagsFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					buildFullFlex(detailText)
					return nil
				}
				if event.Key() == tcell.KeyTab {
					focus := app.GetFocus().(*tview.List)
					if focus == actualLabelList {
						app.SetFocus(labelList)
					} else {
						app.SetFocus(actualLabelList)
					}
				} else if event.Rune() == 63 {
					// ? deck_help menu
					buildHelp(editTagsFlex, deck_help.HelpLabels)
				}
				return event
			})

			buildFullFlex(editTagsFlex)
		} else if event.Rune() == 63 {
			// ? deck_help menu
			buildHelp(detailText, deck_help.HelpView)
		}
		return event
	})

	detailEditText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			detailText.Clear()
			detailText.SetTitle(fmt.Sprintf(" %s ", editableCard.Title))
			detailText.SetText(formatDescription(editableCard.Description))
			go buildFullFlex(detailText)
		} else if event.Key() == tcell.KeyF2 {
			editableCard.Description = detailEditText.GetText()
			go editCard()
			cardsMap[editableCard.Id] = editableCard
			detailText.SetText(formatDescription(editableCard.Description))
			buildFullFlex(detailText)
		}
		return event
	})
	detailText.SetBorder(true)
	detailText.SetBorderColor(utils.GetColor(configuration.Color))

	detailEditText.SetBorder(true)
	detailEditText.SetBorderColor(utils.GetColor(configuration.Color))

	pages.AddPage("Main", fullFlex, true, true)

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func getNextFocus(index int) tview.Primitive {
	if index == len(primitivesIndexMap) {
		index = 0
	}
	return primitivesIndexMap[index]
}

func getCardId(name string) int {
	split := strings.Split(name, "-")
	v := strings.Split(strings.Split(split[0], "]")[1], "[")[0]
	_ = v
	cardId, _ := strconv.Atoi(strings.TrimSpace(v[1:]))
	return cardId
}

func buildStacks() {
	mainFlex.Clear()
	for index, s := range stacks {
		todoList := tview.NewList()
		todoList.SetTitle(fmt.Sprintf(" %s ", s.Title))
		todoList.SetBorder(true)

		todoList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTAB {
				return nil
			}
			if event.Key() == tcell.KeyRight {
				moveCardToStack(*todoList, tcell.KeyRight)
				return nil
			}
			if event.Key() == tcell.KeyLeft {
				moveCardToStack(*todoList, tcell.KeyLeft)
				return nil
			}
			return event
		})

		for _, card := range s.Cards {
			var labels = ""
			for i, label := range card.Labels {
				labels = fmt.Sprintf("%s[#%s]%s[white]", labels, label.Color, label.Title)
				if i != len(card.Labels)-1 {
					labels = fmt.Sprintf("%s, ", labels)
				}
			}
			cardsMap[card.Id] = card
			todoList.AddItem(fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, card.Id, card.Title), labels, rune(0), nil)
		}

		todoList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			cardId := getCardId(name)

			detailText.SetTitle(fmt.Sprintf(" %s ", cardsMap[cardId].Title))
			detailText.SetText(formatDescription(cardsMap[cardId].Description))
			editableCard = cardsMap[cardId]
			buildFullFlex(detailText)
		})

		todoList.SetFocusFunc(func() {
			todoList.SetTitleColor(utils.GetColor(configuration.Color))
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
	if primitive != mainFlex {
		footerBar.SetText("Press [yellow]?[white] for help, [yellow]ESC[white] to go back")

	} else {
		footerBar.SetText("Press [yellow]?[white] for help, [yellow]q[white] to exit")
	}
	app.SetFocus(primitive)
}

func buildSwitchBoard(configuration utils.Configuration) {
	boardList.SetBorder(true)
	boardList.SetBorderColor(utils.GetColor(configuration.Color))
	boardList.SetTitle("Select Boards")
	for _, b := range boards {
		boardList.AddItem(b.Title, "", rune(0), nil)
	}
	boardList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			go buildFullFlex(mainFlex)
		} else if event.Rune() == 63 {
			// ? deck_help menu
			buildHelp(boardList, deck_help.HelpBoards)
		}
		return event
	})
	boardList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
		currentBoard = boards[index]
		mainFlex.SetTitle(fmt.Sprintf(" TUI DECK: [#%s]%s", currentBoard.Color, currentBoard.Title))
		var err error = nil
		stacks, err = deck_http.GetStacks(currentBoard.Id, configuration)
		if err != nil {
			footerBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
		}
		buildStacks()
		go buildFullFlex(mainFlex)
	})
}

func buildHelp(primitive tview.Primitive, helpView *tview.TextView) {
	help := tview.NewFrame(helpView)
	help.SetBorder(true)
	help.SetBorderColor(utils.GetColor(configuration.Color))
	help.SetTitle(deck_help.HelpMain.GetTitle())
	go buildFullFlex(help)

	help.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			go buildFullFlex(primitive)
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

func moveCardToStack(todoList tview.List, key tcell.Key) {
	i := todoList.GetCurrentItem()
	name, _ := todoList.GetItemText(i)
	cardId := getCardId(name)
	card := cardsMap[cardId]

	var position int
	switch key {
	case tcell.KeyLeft:
		if card.StackId == 1 {
			return
		}
		position = card.StackId - 1
		break
	case tcell.KeyRight:
		if card.StackId == len(stacks) {
			return
		}
		position = card.StackId + 1
		break
	}
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"stackId": "%d", "title": "%s", "type": "plain", "owner":"%s"}`,
		position, card.Title, configuration.User), "\n", `\n`)

	_, err := deck_http.UpdateCard(currentBoard.Id, card.StackId, card.Id, jsonBody, configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error moving card: %s", err.Error()))
		return
	}
	stacks, err = deck_http.GetStacks(currentBoard.Id, configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
		return
	}
	buildStacks()
}

func deleteLabel(jsonBody string) {
	_, err := deck_http.DeleteLabel(currentBoard.Id, editableCard.StackId, editableCard.Id, jsonBody, configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error deleting tag from card: %s", err.Error()))
	}
}
func assignLabel(jsonBody string) {
	_, err := deck_http.AssignLabel(currentBoard.Id, editableCard.StackId, editableCard.Id, jsonBody, configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error deleting tag from card: %s", err.Error()))
	}
}

func editCard() {
	jsonBody := strings.ReplaceAll(
		fmt.Sprintf(`{"description": "%s", "title": "%s", "type": "plain", "owner":"%s"}`,
			editableCard.Description, editableCard.Title, configuration.User), "\n", `\n`)
	var err error
	_, err = deck_http.UpdateCard(currentBoard.Id, editableCard.StackId, editableCard.Id, jsonBody, configuration)
	if err != nil {
		footerBar.SetText(fmt.Sprintf("Error updating card: %s", err.Error()))
	}

}

func updateStacks() {
	for i, s := range stacks {
		if s.Id == editableCard.StackId {
			for j, c := range s.Cards {
				if c.Id == editableCard.Id {
					stacks[i].Cards[j] = editableCard
					break
				}
			}
			break
		}
	}
}
