package main

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"time"
	"tui-deck/deck_board"
	"tui-deck/deck_card"
	"tui-deck/deck_comment"
	"tui-deck/deck_db"
	"tui-deck/deck_help"
	"tui-deck/deck_http"
	"tui-deck/deck_stack"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var app = tview.NewApplication()
var pages = tview.NewPages()

var configuration utils.Configuration

func main() {
	deck_help.InitHelp()
	var err error
	configFile, err := utils.InitConfingDirectory()
	if err != nil {
		deck_ui.FooterBar.SetText(err.Error())
	}

	configuration, err = utils.GetConfiguration(configFile)
	if err != nil {
		deck_ui.FooterBar.SetText(err.Error())
	}

	fmt.Print("Getting boards...\n")
	deck_ui.Init(app, configuration)
	deck_board.Init(app, configuration)
	deck_board.Boards, err = deck_http.GetBoards(configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting boards: %s", err.Error()))
	}

	if len(deck_board.Boards) > 0 {
		for i, b := range deck_board.Boards {
			localBoardFile, _ := os.Open(fmt.Sprintf("%s/db/board-%d.json", configuration.ConfigDir, b.Id))
			decoder := json.NewDecoder(localBoardFile)
			localBoard := deck_structs.Board{}
			err = decoder.Decode(&localBoard)
			if b.Etag != localBoard.Etag {
				var boardFile *os.File
				boardFile, err = utils.CreateFile(fmt.Sprintf("%s/db/board-%d.json", configuration.ConfigDir, b.Id))
				if err != nil {
					panic(err)
				}
				marshal, _ := json.Marshal(b)
				_, err = boardFile.Write(marshal)
				if err != nil {
					panic(err)
				}
				b.Updated = true
				deck_board.Boards[i] = b
			}
		}
		fmt.Print("Getting board detail...\n")
		deck_board.CurrentBoard, err = deck_db.GetBoardDetails(deck_board.Boards[0].Id, deck_board.Boards[0].Updated, configuration)
		if err != nil {
			deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting board detail: %s", err.Error()))
		}
		go deck_board.BuildSwitchBoard(configuration)
	} else {
		deck_ui.FooterBar.SetText("No boards found")
	}
	deck_ui.MainFlex.SetTitle(fmt.Sprintf(" TUI DECK: [#%s]%s ", deck_board.CurrentBoard.Color, deck_board.CurrentBoard.Title))
	fmt.Print("Getting stacks...\n")
	deck_stack.Init(app, configuration)
	deck_card.Init(app, configuration, deck_board.CurrentBoard)
	deck_comment.Init(app, configuration)
	deck_stack.Stacks, err = deck_db.GetStacks(deck_board.CurrentBoard.Id, deck_board.CurrentBoard.Updated, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting stacks: %s", err.Error()))
	}
	deck_card.BuildStacks()

	deck_ui.MainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if deck_card.Modal.HasFocus() {
			return deck_card.Modal.GetInputCapture()(event)
		}
		if event.Rune() == 113 {
			// q -> quit app
			app.Stop()
		} else if event.Key() == tcell.KeyTab {
			// tab -> switch focus between stacks
			primitive := app.GetFocus()
			list := primitive.(*tview.List)
			list.SetTitleColor(tcell.ColorWhite)
			actualPrimitiveIndex := deck_ui.Primitives[primitive]
			app.SetFocus(deck_ui.GetNextFocus(actualPrimitiveIndex + 1))
		} else if event.Rune() == 114 {
			// r -> reload stacks
			deck_stack.Stacks, err = deck_http.GetStacks(deck_board.CurrentBoard.Id, configuration)
			if err != nil {
				deck_ui.FooterBar.SetText(fmt.Sprintf("Error reloading stacks: %s", err.Error()))
			}
			deck_card.BuildStacks()
		} else if event.Rune() == 115 {
			// s -> switch board
			deck_ui.BuildFullFlex(deck_board.BoardFlex, nil)
		} else if event.Rune() == 97 {
			// a -> add card
			if len(deck_stack.Stacks) == 0 {
				return nil
			}
			actualList := app.GetFocus().(*tview.List)
			addForm, card := deck_card.BuildAddForm()
			addForm.AddButton("Save", func() {
				dueDate := card.DueDate
				pattern := "02/01/2006 15:04"
				parse, err2 := time.Parse(pattern, dueDate)
				if err2 != nil {
					deck_ui.FooterBar.SetText(fmt.Sprintf("Not a valid date, format must be dd/MM/YYYY HH:mm: %s", err2.Error()))
					return
				}
				card.DueDate = parse.Format("2006-01-02T15:04:05+00:00")
				deck_card.AddCard(actualList, *card)
			})
			deck_ui.BuildFullFlex(addForm, nil)
		} else if event.Rune() == 100 {
			if len(deck_stack.Stacks) == 0 {
				return nil
			}
			// d -> delete card
			actualList := app.GetFocus().(*tview.List)
			var _, stack, _ = deck_stack.GetActualStack(actualList)
			var currentItemIndex = actualList.GetCurrentItem()
			mainText, _ := actualList.GetItemText(currentItemIndex)
			cardId := utils.GetId(mainText)
			deck_card.DeleteCard(cardId, stack, actualList, currentItemIndex)

		} else if event.Key() == tcell.KeyCtrlA {
			// ctrl + a -> add stack
			addForm, stack := deck_stack.BuildAddForm(deck_structs.Stack{})
			addForm.AddButton("Save", func() {
				err := deck_stack.AddStack(deck_board.CurrentBoard.Id, *stack)
				deck_card.BuildStacks()
				deck_ui.BuildFullFlex(deck_ui.MainFlex, err)
			})
			deck_ui.BuildFullFlex(addForm, nil)

		} else if event.Key() == tcell.KeyCtrlD {
			// ctrl + d -> delete stack
			if len(deck_stack.Stacks) == 0 {
				return nil
			}

			actualList := app.GetFocus().(*tview.List)
			index := deck_ui.Primitives[app.GetFocus()]
			currentStack := deck_stack.Stacks[index]

			deck_stack.DeleteStack(currentStack.Id, actualList)
			deck_stack.Modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					go func() {
						_, err = deck_http.DeleteStack(deck_board.CurrentBoard.Id, currentStack.Id, configuration)
						if err != nil {
							deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting stack: %s", err.Error()))
						}
					}()
					deck_ui.MainFlex.RemoveItem(deck_stack.Modal)
					deck_ui.MainFlex.RemoveItem(actualList)
					deck_stack.Stacks = append(deck_stack.Stacks[:index], deck_stack.Stacks[index+1:]...)
					deck_card.BuildStacks()
					deck_ui.BuildFullFlex(deck_ui.MainFlex, nil)
				} else if buttonLabel == "No" {
					deck_ui.MainFlex.RemoveItem(deck_stack.Modal)
					app.SetFocus(actualList)
				}
			})

		} else if event.Key() == tcell.KeyCtrlE {
			// ctrl + e -> edit stack
			actualList := app.GetFocus().(*tview.List)

			index := deck_ui.Primitives[app.GetFocus()]
			currentStack := deck_stack.Stacks[index]
			editForm, editedStack := deck_stack.BuildAddForm(currentStack)
			editForm.AddButton("Save", func() {
				actualList.SetTitle(fmt.Sprintf("# %s ", editedStack.Title))
				var err error
				go func() {
					err = deck_stack.EditStack(deck_board.CurrentBoard.Id, *editedStack)
				}()

				deck_stack.Stacks[index] = *editedStack
				deck_card.BuildStacks()
				deck_ui.BuildFullFlex(deck_ui.MainFlex, err)
			})
			deck_ui.BuildFullFlex(editForm, nil)

		} else if event.Rune() == 63 {
			// ? deck_help menu
			deck_ui.BuildHelp(deck_ui.MainFlex, deck_help.HelpMain)
		}

		return event
	})
	deck_card.BuildCardViewer()
	pages.AddPage("Main", deck_ui.FullFlex, true, true)
	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}
