package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strings"
	"tui-deck/deck_db"
	"tui-deck/deck_help"
	"tui-deck/deck_http"
	"tui-deck/deck_markdown"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var app = tview.NewApplication()
var pages = tview.NewPages()

var boardFlex = tview.NewFlex()
var detailText = tview.NewTextView()
var detailEditText = tview.NewTextArea()
var boardList = tview.NewList()
var editTagsFlex = tview.NewFlex()
var modal = tview.NewModal()

var stacks []deck_structs.Stack
var boards []deck_structs.Board
var cardsMap = make(map[int]deck_structs.Card)
var editableCard = deck_structs.Card{}
var currentBoard deck_structs.Board

var primitives = make(map[tview.Primitive]int)
var primitivesIndexMap = make(map[int]tview.Primitive)

var configuration utils.Configuration
var configDir string
var configFile string

func main() {
	deck_help.InitHelp()
	var err error
	configFile, configDir, err = utils.InitConfingDirectory()
	if err != nil {
		deck_ui.FooterBar.SetText(err.Error())
	}

	configuration, err = utils.GetConfiguration(configFile)
	if err != nil {
		deck_ui.FooterBar.SetText(err.Error())
	}

	fmt.Print("Getting boards...\n")
	boards, err = deck_http.GetBoards(configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting boards: %s", err.Error()))
	}

	if len(boards) > 0 {
		for i, b := range boards {
			localBoardFile, _ := os.Open(fmt.Sprintf("%s/db/board-%d.json", configDir, b.Id))
			decoder := json.NewDecoder(localBoardFile)
			localBoard := deck_structs.Board{}
			err = decoder.Decode(&localBoard)
			if b.Etag != localBoard.Etag {
				var boardFile *os.File
				boardFile, err = utils.CreateFile(fmt.Sprintf("%s/db/board-%d.json", configDir, b.Id))
				if err != nil {
					panic(err)
				}
				marshal, _ := json.Marshal(b)
				_, err = boardFile.Write(marshal)
				if err != nil {
					panic(err)
				}
				b.Updated = true
				boards[i] = b
			}
		}
		fmt.Print("Getting board detail...\n")
		currentBoard, err = deck_db.GetBoardDetails(boards[0].Id, boards[0].Updated, configDir, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting board detail: %s", err.Error()))

		}
		go buildSwitchBoard(configuration)
	} else {
		deck_ui.FooterBar.SetText("No boards found")
	}

	deck_ui.Init(app, configuration, currentBoard)

	fmt.Print("Getting stacks...\n")
	stacks, err = deck_db.GetStacks(currentBoard.Id, currentBoard.Updated, configDir, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
	}

	buildStacks()

	deck_ui.MainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if modal.HasFocus() {
			return modal.GetInputCapture()(event)
		}

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
				deck_ui.FooterBar.SetText(fmt.Sprintf("Error reloading stacks: %s", err.Error()))
			}
			buildStacks()
		} else if event.Rune() == 115 {
			// s -> switch board
			boardFlex.Clear()
			boardFlex.AddItem(boardList, 0, 1, true)
			deck_ui.BuildFullFlex(boardFlex)

		} else if event.Rune() == 97 {
			// a -> add card
			actualList := app.GetFocus().(*tview.List)

			addForm, card := deck_ui.BuildAddForm()

			//TODO add due Date input field
			addForm.AddButton("Save", func() {
				addCard(*actualList, *card)

			})

			deck_ui.BuildFullFlex(addForm)

		} else if event.Rune() == 100 {
			// d -> delete card
			actualList := app.GetFocus().(*tview.List)
			var _, stack, _ = getActualStack(*actualList)

			var currentItemIndex = actualList.GetCurrentItem()
			mainText, _ := actualList.GetItemText(currentItemIndex)
			cardId := utils.GetId(mainText)

			modal.ClearButtons()
			modal.SetText(fmt.Sprintf("Are you sure to delete card #%d?", cardId))
			modal.SetBackgroundColor(utils.GetColor(configuration.Color))

			modal.AddButtons([]string{"Yes", "No"})

			modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEnter {
					return event
				}
				return nil
			})

			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					go func() {
						_, _ = deck_http.DeleteCard(currentBoard.Id, stack.Id, cardId, configuration)
					}()
					actualList.RemoveItem(currentItemIndex)
					deck_ui.MainFlex.RemoveItem(modal)
					app.SetFocus(actualList)
				} else if buttonLabel == "No" {
					deck_ui.MainFlex.RemoveItem(modal)
					app.SetFocus(actualList)
				}
			})

			deck_ui.MainFlex.AddItem(modal, 0, 0, false)
			app.SetFocus(modal)

		} else if event.Rune() == 63 {
			// ? deck_help menu
			deck_ui.BuildHelp(deck_ui.MainFlex, deck_help.HelpMain)
		}

		return event

	})

	detailText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// ESC -> back to main view
			deck_ui.BuildFullFlex(deck_ui.MainFlex)
		} else if event.Rune() == 101 {
			// e -> edit description
			detailEditText.SetTitle(fmt.Sprintf(" %s- EDIT", detailText.GetTitle()))
			detailEditText.SetText(utils.FormatDescription(editableCard.Description), true)
			deck_ui.BuildFullFlex(detailEditText)
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
					deck_ui.BuildFullFlex(detailText)
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
					deck_ui.BuildHelp(editTagsFlex, deck_help.HelpLabels)
				}
				return event
			})

			deck_ui.BuildFullFlex(editTagsFlex)
		} else if event.Rune() == 63 {
			// ? -> deck_help menu
			deck_ui.BuildHelp(detailText, deck_help.HelpView)
		}
		return event
	})

	detailEditText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			detailText.Clear()
			detailText.SetTitle(fmt.Sprintf(" %s ", editableCard.Title))
			detailText.SetText(deck_markdown.GetMarkDownDescription(utils.FormatDescription(editableCard.Description), configuration))
			deck_ui.BuildFullFlex(detailText)
		} else if event.Key() == tcell.KeyF2 {
			editableCard.Description = detailEditText.GetText()
			go editCard()
			cardsMap[editableCard.Id] = editableCard
			detailText.SetText(deck_markdown.GetMarkDownDescription(utils.FormatDescription(editableCard.Description), configuration))
			deck_ui.BuildFullFlex(detailText)
		}
		return event
	})
	detailText.SetBorder(true)
	detailText.SetBorderColor(utils.GetColor(configuration.Color))

	detailEditText.SetBorder(true)
	detailEditText.SetBorderColor(utils.GetColor(configuration.Color))

	pages.AddPage("Main", deck_ui.FullFlex, true, true)

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func getActualStack(actualList tview.List) (int, deck_structs.Stack, error) {
	for i, s := range stacks {
		if s.Title == strings.TrimSpace(actualList.GetTitle()) {
			return i, s, nil
		}
	}
	return 0, deck_structs.Stack{}, errors.New("not found")
}

func getNextFocus(index int) tview.Primitive {
	if index == len(primitivesIndexMap) {
		index = 0
	}
	return primitivesIndexMap[index]
}

func buildStacks() {
	deck_ui.MainFlex.Clear()
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
			var labels = utils.BuildLabels(card)
			cardsMap[card.Id] = card
			todoList.AddItem(fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, card.Id, card.Title), labels, rune(0), nil)
		}

		todoList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			cardId := utils.GetId(name)

			detailText.SetTitle(fmt.Sprintf(" %s ", cardsMap[cardId].Title))
			detailText.SetDynamicColors(true)

			description := utils.FormatDescription(cardsMap[cardId].Description)
			detailText.SetText(deck_markdown.GetMarkDownDescription(description, configuration))
			editableCard = cardsMap[cardId]
			deck_ui.BuildFullFlex(detailText)
		})

		todoList.SetFocusFunc(func() {
			todoList.SetTitleColor(utils.GetColor(configuration.Color))
		})

		primitives[todoList] = index
		primitivesIndexMap[index] = todoList

		deck_ui.MainFlex.AddItem(todoList, 0, 1, true)
		primitive := deck_ui.MainFlex.GetItem(0)
		app.SetFocus(primitive)
	}
}

func buildSwitchBoard(configuration utils.Configuration) {
	boardList.SetBorder(true)
	boardList.SetBorderColor(utils.GetColor(configuration.Color))
	boardList.SetTitle("Select Boards")
	for _, b := range boards {
		boardList.AddItem(fmt.Sprintf("[#%s]#%d - %s", b.Color, b.Id, b.Title), "", rune(0), nil)
	}
	boardList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			deck_ui.BuildFullFlex(deck_ui.MainFlex)
		} else if event.Rune() == 97 {
			// a -> add board
			addForm, board := deck_ui.BuildAddBoardForm(deck_structs.Board{})
			addForm.AddButton("Save", func() {
				addBoard(*board)
			})

			deck_ui.BuildFullFlex(addForm)
		} else if event.Rune() == 101 {
			// e -> edit board
			selectedBoardIndex := boardList.GetCurrentItem()
			text, _ := boardList.GetItemText(selectedBoardIndex)

			boardId := utils.GetId(text)
			board := deck_structs.Board{}
			for _, b := range boards {
				if b.Id == boardId {
					board = b
					break
				}
			}

			editForm, editedBoard := deck_ui.BuildAddBoardForm(board)
			editForm.AddButton("Save", func() {
				go editBoard(*editedBoard)
				boardList.SetItemText(selectedBoardIndex, fmt.Sprintf("[#%s]#%d - %s", editedBoard.Color, editedBoard.Id, editedBoard.Title), "")
				for i, b := range boards {
					if b.Id == editedBoard.Id {
						boards[i] = *editedBoard
						break
					}
				}
				deck_ui.BuildFullFlex(boardFlex)
			})
			deck_ui.BuildFullFlex(editForm)

		} else if event.Rune() == 100 {
			// d -> delete board
			selectedBoardIndex := boardList.GetCurrentItem()
			text, _ := boardList.GetItemText(selectedBoardIndex)
			boardId := utils.GetId(text)
			modal = tview.NewModal()
			modal.ClearButtons()
			modal.SetText(fmt.Sprintf("Are you sure to delete bord #%d?", boardId))
			modal.SetBackgroundColor(utils.GetColor(configuration.Color))
			modal.AddButtons([]string{"Yes", "No"})

			modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEnter {
					return event
				}
				return nil
			})

			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					go func() {
						_, _ = deck_http.DeleteBoard(boardId, configuration)
					}()
					boardList.RemoveItem(selectedBoardIndex)
					boardFlex.RemoveItem(modal)
					app.SetFocus(boardList)
				} else if buttonLabel == "No" {
					boardFlex.RemoveItem(modal)
					app.SetFocus(boardList)
				}
			})

			boardFlex.AddItem(modal, 0, 0, false)
			app.SetFocus(modal)

		} else if event.Rune() == 63 {
			// ? deck_help menu
			deck_ui.BuildHelp(boardList, deck_help.HelpBoards)
		}
		return event
	})
	boardList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
		var err error
		currentBoard, err = deck_db.GetBoardDetails(boards[index].Id, boards[index].Updated, configDir, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting board detail: %s", err.Error()))

		}
		deck_ui.MainFlex.SetTitle(fmt.Sprintf(" TUI DECK: [#%s]%s", currentBoard.Color, currentBoard.Title))

		stacks, err = deck_db.GetStacks(currentBoard.Id, boards[index].Updated, configDir, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
		}
		buildStacks()
		deck_ui.BuildFullFlex(deck_ui.MainFlex)
	})
}

func moveCardToStack(todoList tview.List, key tcell.Key) {
	i := todoList.GetCurrentItem()
	name, _ := todoList.GetItemText(i)
	cardId := utils.GetId(name)
	card := cardsMap[cardId]

	primitive := app.GetFocus()
	actualPrimitiveIndex := primitives[primitive]

	var position int
	var operator int

	switch key {
	case tcell.KeyLeft:
		if card.StackId == 1 {
			return
		}
		position = card.StackId - 1
		operator = -1

		break
	case tcell.KeyRight:
		if card.StackId == len(stacks) {
			return
		}
		position = card.StackId + 1
		operator = 1
		break
	}
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"stackId": "%d", "title": "%s", "type": "plain", "owner":"%s"}`,
		position, card.Title, configuration.User), "\n", `\n`)

	go updateCard(currentBoard.Id, card.StackId, card.Id, jsonBody)

	var labels = utils.BuildLabels(card)
	card.StackId = position
	cardsMap[card.Id] = card
	destList := getNextFocus(actualPrimitiveIndex + operator).(*tview.List)
	destList.InsertItem(0, fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, card.Id, card.Title), labels, rune(0), nil)
	todoList.RemoveItem(i)
}
func updateCard(boardId, stackId int, cardId int, jsonBody string) {
	_, err := deck_http.UpdateCard(boardId, stackId, cardId, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error moving card: %s", err.Error()))
		return
	}
}

func deleteLabel(jsonBody string) {
	_, err := deck_http.DeleteLabel(currentBoard.Id, editableCard.StackId, editableCard.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting tag from card: %s", err.Error()))
	}
}
func assignLabel(jsonBody string) {
	_, err := deck_http.AssignLabel(currentBoard.Id, editableCard.StackId, editableCard.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting tag from card: %s", err.Error()))
	}
}

func editCard() {
	description := strings.ReplaceAll(editableCard.Description, "\"", "\\\"")
	title := strings.ReplaceAll(editableCard.Title, "\"", "\\\"")

	jsonBody := strings.ReplaceAll(
		fmt.Sprintf(`{"description": "%s", "title": "%s", "type": "plain", "owner":"%s"}`,
			description, title, configuration.User), "\n", `\n`)
	var err error
	_, err = deck_http.UpdateCard(currentBoard.Id, editableCard.StackId, editableCard.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error updating card: %s", err.Error()))
	}
}

func addCard(actualList tview.List, card deck_structs.Card) {
	var stackIndex, stack, _ = getActualStack(actualList)

	jsonBody := fmt.Sprintf(`{"title":"%s", "description": "%s", "type": "plain", "order": 0}`, card.Title, card.Description)
	var newCard deck_structs.Card
	var err error
	newCard, err = deck_http.AddCard(currentBoard.Id, stack.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
	}

	actualList.InsertItem(0, fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, newCard.Id, newCard.Title), "", rune(0), nil)
	cardsMap[newCard.Id] = newCard
	detailText.Clear()
	editableCard = newCard
	stacks[stackIndex].Cards = append(stacks[stackIndex].Cards[:1], stacks[stackIndex].Cards[0:]...)
	stacks[stackIndex].Cards[0] = newCard
	detailText.SetTitle(fmt.Sprintf(" %s ", newCard.Title))
	detailText.SetText(utils.FormatDescription(newCard.Description))
	deck_ui.BuildFullFlex(detailText)
}

func addBoard(board deck_structs.Board) {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, board.Title, board.Color)
	var newBoard deck_structs.Board
	var err error
	newBoard, err = deck_http.AddBoard(jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
	}
	boards = append(boards, newBoard)
	boardList.AddItem(fmt.Sprintf("[#%s]#%d - %s", newBoard.Color, newBoard.Id, newBoard.Title), "", rune(0), nil)

	deck_ui.BuildFullFlex(boardFlex)
}

func editBoard(board deck_structs.Board) {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, board.Title, board.Color)
	var err error
	_, err = deck_http.EditBoard(board.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
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
