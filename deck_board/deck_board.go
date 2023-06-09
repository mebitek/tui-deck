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
var EditTagsFlex *tview.Flex
var modal = tview.NewModal()

var Boards []deck_structs.Board
var CurrentBoard deck_structs.Board

var app *tview.Application
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration) {
	BoardFlex = tview.NewFlex()
	BoardList = tview.NewList()
	EditTagsFlex = tview.NewFlex()

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
			deck_ui.BuildFullFlex(deck_ui.MainFlex, nil)
		} else if event.Rune() == 97 {
			// a -> add board
			addForm, board := buildAddBoardForm(deck_structs.Board{})
			addForm.AddButton("Save", func() {
				addBoard(*board)
			})

			deck_ui.BuildFullFlex(addForm, nil)
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

			editForm, editedBoard := buildAddBoardForm(board)
			editForm.AddButton("Save", func() {
				var err error
				go func() {
					err = editBoard(*editedBoard)
				}()
				BoardList.SetItemText(selectedBoardIndex, fmt.Sprintf("[#%s]#%d - %s", editedBoard.Color, editedBoard.Id, editedBoard.Title), "")
				for i, b := range Boards {
					if b.Id == editedBoard.Id {
						Boards[i] = *editedBoard
						break
					}
				}
				deck_ui.BuildFullFlex(BoardFlex, err)
			})
			deck_ui.BuildFullFlex(editForm, nil)
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
				if event.Key() == tcell.KeyEsc {
					BoardFlex.RemoveItem(modal)
					app.SetFocus(BoardList)
				}
				if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEnter {
					return event
				}
				return nil
			})

			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					go func() {
						_, err := deck_http.DeleteBoard(boardId, configuration)
						if err != nil {
							deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleteing board: %s", err.Error()))
						}
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

		} else if event.Rune() == 116 {
			// t -> tags
			currentIndex := BoardList.GetCurrentItem()
			text, _ := BoardList.GetItemText(currentIndex)

			boardId := utils.GetId(text)

			board, _ := deck_db.GetBoardDetails(boardId, Boards[currentIndex].Updated, configuration)

			EditTagsFlex.Clear()
			actualLabelList := tview.NewList()
			actualLabelList.SetBorder(true)
			actualLabelList.SetTitle(" delete labels ")
			actualLabelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyTab {
					return nil
				}
				return event
			})
			for _, label := range board.Labels {
				actualLabelList.AddItem(fmt.Sprintf("[#%s]#%d - %s", label.Color, label.Id, label.Title), "",
					rune(0), nil)
			}
			actualLabelList.SetSelectedFunc(func(index int, name string, secondName string, rune rune) {

				labelId := utils.GetId(name)

				go DeleteLabel(boardId, labelId)
				board.Updated = true
				board.Labels = append(board.Labels[:index], board.Labels[index+1:]...)
				for i, b := range Boards {
					if b.Id == boardId {
						Boards[i] = board
						break
					}
				}
				actualLabelList.RemoveItem(index)

				app.SetFocus(actualLabelList)
			})

			actualLabelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Rune() == 97 {
					// a -> add label
					addForm, label := buildAddLabelForm(deck_structs.Label{})
					addForm.AddButton("Save", func() {
						board.Updated = true
						addLabel(*label, &board, actualLabelList)
					})

					deck_ui.BuildFullFlex(addForm, nil)
				} else if event.Rune() == 101 {
					// e -> edit label

					selectedLabelIndex := actualLabelList.GetCurrentItem()
					labelText, _ := actualLabelList.GetItemText(selectedLabelIndex)

					labelId := utils.GetId(labelText)
					label := deck_structs.Label{}
					for _, l := range board.Labels {
						if l.Id == labelId {
							label = l
							break
						}
					}

					editForm, editedLabel := buildAddLabelForm(label)
					editForm.AddButton("Save", func() {
						board.Updated = true
						var err error
						go func() {
							err = editLabel(boardId, *editedLabel)
						}()
						actualLabelList.SetItemText(selectedLabelIndex, fmt.Sprintf("[#%s]#%d - %s", editedLabel.Color, editedLabel.Id, editedLabel.Title), "")
						for i, l := range board.Labels {
							if l.Id == editedLabel.Id {
								board.Labels[i] = *editedLabel
								break
							}
						}
						deck_ui.BuildFullFlex(EditTagsFlex, err)
					})
					deck_ui.BuildFullFlex(editForm, nil)
				}
				return event
			})

			EditTagsFlex.SetDirection(tview.FlexColumn)
			EditTagsFlex.SetBorder(true)
			EditTagsFlex.SetBorderColor(utils.GetColor(configuration.Color))
			EditTagsFlex.SetTitle(fmt.Sprintf(" [#%s]%s[-:-:-] - EDIT TAGS ", board.Color, board.Title))
			EditTagsFlex.AddItem(actualLabelList, 0, 1, true)
			EditTagsFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					deck_ui.BuildFullFlex(BoardFlex, nil)
					return nil
				} else if event.Rune() == 63 {
					// ? deck_help menu
					deck_ui.BuildHelp(EditTagsFlex, deck_help.HelpLabels)
				}
				return event
			})
			deck_ui.BuildFullFlex(EditTagsFlex, nil)

		} else if event.Rune() == 63 {
			// ? deck_help menu
			deck_ui.BuildHelp(BoardList, deck_help.HelpBoards)
		}
		return event
	})
	BoardList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
		var err error
		CurrentBoard, err = deck_db.GetBoardDetails(Boards[index].Id, Boards[index].Updated, configuration)
		Boards[index] = CurrentBoard
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting board detail: %s", err.Error()))

		}
		deck_ui.MainFlex.SetTitle(fmt.Sprintf(" TUI DECK: [#%s]%s", CurrentBoard.Color, CurrentBoard.Title))

		deck_stack.Stacks, err = deck_db.GetStacks(CurrentBoard.Id, Boards[index].Updated, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
		}
		deck_card.BuildStacks()
		deck_ui.BuildFullFlex(deck_ui.MainFlex, err)
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

	deck_ui.BuildFullFlex(BoardFlex, err)
}

func editBoard(board deck_structs.Board) error {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, board.Title, board.Color)
	var err error
	_, err = deck_http.EditBoard(board.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
		return err
	}
	return nil
}

func DeleteLabel(boardId int, labelId int) {
	_, err := deck_http.DeleteBoardLabel(boardId, labelId, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting tag from board: %s", err.Error()))
	}
}

func addLabel(label deck_structs.Label, board *deck_structs.Board, actualLabelList *tview.List) {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, label.Title, label.Color)
	var newLabel deck_structs.Label
	var err error
	newLabel, err = deck_http.AddBoardLabel(board.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
		return
	}
	board.Labels = append(board.Labels, newLabel)

	for i, b := range Boards {
		if b.Id == board.Id {
			Boards[i] = *board
			break
		}
	}

	actualLabelList.AddItem(fmt.Sprintf("[#%s]#%d - %s", newLabel.Color, newLabel.Id, newLabel.Title), "", rune(0), nil)

	deck_ui.BuildFullFlex(EditTagsFlex, err)
}

func editLabel(boardId int, label deck_structs.Label) error {
	jsonBody := fmt.Sprintf(`{"title":"%s", "color": "%s"}`, label.Title, label.Color)
	var err error
	_, err = deck_http.EditBoardLabel(boardId, label.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
		return err
	}
	return nil
}

func buildAddLabelForm(l deck_structs.Label) (*tview.Form, *deck_structs.Label) {
	addForm := tview.NewForm()
	var label = deck_structs.Label{}
	var title = " Add Label "
	if l.Id != 0 {
		label = l
		title = " Edit Label"
	}
	addForm.SetTitle(title)
	addForm.SetBorder(true)
	addForm.SetBorderColor(utils.GetColor(configuration.Color))
	addForm.SetButtonBackgroundColor(utils.GetColor(configuration.Color))
	addForm.SetFieldBackgroundColor(tcell.ColorWhite)
	addForm.SetFieldTextColor(tcell.ColorBlack)
	addForm.SetLabelColor(utils.GetColor(configuration.Color))
	addForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			deck_ui.BuildFullFlex(EditTagsFlex, nil)
			return nil
		}
		return event
	})
	addForm.AddInputField("Title", l.Title, 20, nil, func(title string) {
		label.Title = title
	})
	addForm.AddInputField("Color", l.Color, 20, nil, func(color string) {
		label.Color = color
	})

	return addForm, &label
}

func buildAddBoardForm(b deck_structs.Board) (*tview.Form, *deck_structs.Board) {
	addForm := tview.NewForm()
	var board deck_structs.Board = deck_structs.Board{}
	var title = " Add Board "
	if b.Id != 0 {
		board = b
		title = " Edit Boasrd "
	}
	addForm.SetTitle(title)
	addForm.SetBorder(true)
	addForm.SetBorderColor(utils.GetColor(configuration.Color))
	addForm.SetButtonBackgroundColor(utils.GetColor(configuration.Color))
	addForm.SetFieldBackgroundColor(tcell.ColorWhite)
	addForm.SetFieldTextColor(tcell.ColorBlack)
	addForm.SetLabelColor(utils.GetColor(configuration.Color))
	addForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			deck_ui.BuildFullFlex(deck_ui.MainFlex, nil)
			return nil
		}
		return event
	})
	addForm.AddInputField("Title", b.Title, 20, nil, func(title string) {
		board.Title = title
	})
	addForm.AddInputField("Color", b.Color, 20, nil, func(color string) {
		board.Color = color
	})

	return addForm, &board
}
