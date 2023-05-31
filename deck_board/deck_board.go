package deck_board

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tui-deck/deck_card"
	"tui-deck/deck_db"
	"tui-deck/deck_help"
	"tui-deck/deck_http"
	"tui-deck/deck_stack"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var BoardFlex *tview.Flex
var BoardList *tview.List
var modal = tview.NewModal()

var Boards []deck_structs.Board
var CurrentBoard deck_structs.Board

var app *tview.Application
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration) {
	BoardFlex = tview.NewFlex()
	BoardList = tview.NewList()

	app = application
	configuration = conf

	BoardFlex.Clear()
	BoardFlex.AddItem(BoardList, 0, 1, true)
}

func BuildSwitchBoard(configuration utils.Configuration) {
	BoardList.SetBorder(true)
	BoardList.SetBorderColor(utils.GetColor(configuration.Color))
	BoardList.SetTitle("Select Boards")
	for _, b := range Boards {
		BoardList.AddItem(fmt.Sprintf("[#%s]#%d - %s", b.Color, b.Id, b.Title), "", rune(0), nil)
	}
	BoardList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			selectedBoardIndex := BoardList.GetCurrentItem()
			text, _ := BoardList.GetItemText(selectedBoardIndex)

			boardId := utils.GetId(text)
			board := deck_structs.Board{}
			for _, b := range Boards {
				if b.Id == boardId {
					board = b
					break
				}
			}

			editForm, editedBoard := deck_ui.BuildAddBoardForm(board)
			editForm.AddButton("Save", func() {
				go editBoard(*editedBoard)
				BoardList.SetItemText(selectedBoardIndex, fmt.Sprintf("[#%s]#%d - %s", editedBoard.Color, editedBoard.Id, editedBoard.Title), "")
				for i, b := range Boards {
					if b.Id == editedBoard.Id {
						Boards[i] = *editedBoard
						break
					}
				}
				deck_ui.BuildFullFlex(BoardFlex)
			})
			deck_ui.BuildFullFlex(editForm)
		} else if event.Rune() == 100 {
			// d -> delete board
			selectedBoardIndex := BoardList.GetCurrentItem()
			text, _ := BoardList.GetItemText(selectedBoardIndex)
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
					BoardList.RemoveItem(selectedBoardIndex)
					BoardFlex.RemoveItem(modal)
					app.SetFocus(BoardList)
				} else if buttonLabel == "No" {
					BoardFlex.RemoveItem(modal)
					app.SetFocus(BoardList)
				}
			})

			BoardFlex.AddItem(modal, 0, 0, false)
			app.SetFocus(modal)

		} else if event.Rune() == 63 {
			// ? deck_help menu
			deck_ui.BuildHelp(BoardList, deck_help.HelpBoards)
		}
		return event
	})
	BoardList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
		var err error
		CurrentBoard, err = deck_db.GetBoardDetails(Boards[index].Id, Boards[index].Updated, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting board detail: %s", err.Error()))

		}
		deck_ui.MainFlex.SetTitle(fmt.Sprintf(" TUI DECK: [#%s]%s", CurrentBoard.Color, CurrentBoard.Title))

		deck_stack.Stacks, err = deck_db.GetStacks(CurrentBoard.Id, Boards[index].Updated, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
		}
		deck_card.BuildStacks()
		deck_ui.BuildFullFlex(deck_ui.MainFlex)
	})
}

func addBoard(board deck_structs.Board) {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, board.Title, board.Color)
	var newBoard deck_structs.Board
	var err error
	newBoard, err = deck_http.AddBoard(jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
	}
	Boards = append(Boards, newBoard)
	BoardList.AddItem(fmt.Sprintf("[#%s]#%d - %s", newBoard.Color, newBoard.Id, newBoard.Title), "", rune(0), nil)

	deck_ui.BuildFullFlex(BoardFlex)
}

func editBoard(board deck_structs.Board) {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, board.Title, board.Color)
	var err error
	_, err = deck_http.EditBoard(board.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
	}
}
