package main

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"tui-deck/deck_board"
	"tui-deck/deck_card"
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
			deck_ui.BuildFullFlex(deck_board.BoardFlex)
		} else if event.Rune() == 97 {
			// a -> add card
			actualList := app.GetFocus().(*tview.List)
			addForm, card := deck_ui.BuildAddForm()
			//TODO add due Date input field
			addForm.AddButton("Save", func() {
				deck_card.AddCard(*actualList, *card)
			})
			deck_ui.BuildFullFlex(addForm)
		} else if event.Rune() == 100 {
			// d -> delete card
			actualList := app.GetFocus().(*tview.List)
			var _, stack, _ = deck_stack.GetActualStack(*actualList)
			var currentItemIndex = actualList.GetCurrentItem()
			mainText, _ := actualList.GetItemText(currentItemIndex)
			cardId := utils.GetId(mainText)
			deck_card.DeleteCard(cardId, stack, actualList, currentItemIndex)
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
