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
	"tui-deck/deck_markdown"
	"tui-deck/deck_stack"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var app = tview.NewApplication()
var pages = tview.NewPages()

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
	deck_ui.Init(app, configuration)
	deck_board.Init(app, configuration)
	deck_board.Boards, err = deck_http.GetBoards(configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting boards: %s", err.Error()))
	}

	if len(deck_board.Boards) > 0 {
		for i, b := range deck_board.Boards {
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
				deck_board.Boards[i] = b
			}
		}
		fmt.Print("Getting board detail...\n")
		deck_board.CurrentBoard, err = deck_db.GetBoardDetails(deck_board.Boards[0].Id, deck_board.Boards[0].Updated, configDir, configuration)
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
	deck_stack.Stacks, err = deck_db.GetStacks(deck_board.CurrentBoard.Id, deck_board.CurrentBoard.Updated, configDir, configuration)
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

			deck_card.Modal.ClearButtons()
			deck_card.Modal.SetText(fmt.Sprintf("Are you sure to delete card #%d?", cardId))
			deck_card.Modal.SetBackgroundColor(utils.GetColor(configuration.Color))

			deck_card.Modal.AddButtons([]string{"Yes", "No"})

			deck_card.Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEnter {
					return event
				}
				return nil
			})

			deck_card.Modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					go func() {
						_, _ = deck_http.DeleteCard(deck_board.CurrentBoard.Id, stack.Id, cardId, configuration)
					}()
					actualList.RemoveItem(currentItemIndex)
					deck_ui.MainFlex.RemoveItem(deck_card.Modal)
					app.SetFocus(actualList)
				} else if buttonLabel == "No" {
					deck_ui.MainFlex.RemoveItem(deck_card.Modal)
					app.SetFocus(actualList)
				}
			})

			deck_ui.MainFlex.AddItem(deck_card.Modal, 0, 0, false)
			app.SetFocus(deck_card.Modal)

		} else if event.Rune() == 63 {
			// ? deck_help menu
			deck_ui.BuildHelp(deck_ui.MainFlex, deck_help.HelpMain)
		}

		return event

	})

	deck_card.DetailText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// ESC -> back to main view
			deck_ui.BuildFullFlex(deck_ui.MainFlex)
		} else if event.Rune() == 101 {
			// e -> edit description
			deck_card.DetailEditText.SetTitle(fmt.Sprintf(" %s- EDIT", deck_card.DetailText.GetTitle()))
			deck_card.DetailEditText.SetText(utils.FormatDescription(deck_card.EditableCard.Description), true)
			deck_ui.BuildFullFlex(deck_card.DetailEditText)
		} else if event.Rune() == 116 {
			// t -> tags
			deck_card.EditTagsFlex.Clear()
			actualLabelList := tview.NewList()
			actualLabelList.SetBorder(true)
			actualLabelList.SetTitle(" delete labels ")
			actualLabelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyTab {
					return nil
				}
				return event
			})
			for _, label := range deck_card.EditableCard.Labels {
				actualLabelList.AddItem(fmt.Sprintf("[#%s]%s", label.Color, label.Title), "",
					rune(0), nil)
			}
			actualLabelList.SetSelectedFunc(func(index int, name string, secondName string, rune rune) {
				label := deck_card.EditableCard.Labels[index]
				jsonBody := fmt.Sprintf(`{"labelId": %d}`, label.Id)
				go deck_card.DeleteLabel(jsonBody)
				deck_card.EditableCard.Labels = append(deck_card.EditableCard.Labels[:index], deck_card.EditableCard.Labels[index+1:]...)
				deck_card.CardsMap[deck_card.EditableCard.Id] = deck_card.EditableCard
				actualLabelList.RemoveItem(index)

				updateStacks()
				deck_card.BuildStacks()
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
			for _, label := range deck_board.CurrentBoard.Labels {
				labelList.AddItem(fmt.Sprintf("[#%s]%s", label.Color, label.Title), "",
					rune(0), nil)
			}

			labelList.SetSelectedFunc(func(index int, name string, secondName string, rune rune) {
				label := deck_board.CurrentBoard.Labels[index]
				jsonBody := fmt.Sprintf(`{"labelId": %d }`, label.Id)
				go deck_card.AssignLabel(jsonBody)
				deck_card.EditableCard.Labels = append(deck_card.EditableCard.Labels, label)
				deck_card.CardsMap[deck_card.EditableCard.Id] = deck_card.EditableCard
				actualLabelList.AddItem(fmt.Sprintf("[#%s]%s", label.Color, label.Title), "",
					rune, nil)
				updateStacks()
				deck_card.BuildStacks()
				app.SetFocus(labelList)
			})

			deck_card.EditTagsFlex.SetDirection(tview.FlexColumn)
			deck_card.EditTagsFlex.SetBorder(true)
			deck_card.EditTagsFlex.SetBorderColor(utils.GetColor(configuration.Color))
			deck_card.EditTagsFlex.SetTitle(fmt.Sprintf(" %s- EDIT TAGS", deck_card.DetailText.GetTitle()))

			deck_card.EditTagsFlex.AddItem(actualLabelList, 0, 1, true)
			deck_card.EditTagsFlex.AddItem(labelList, 0, 1, true)
			deck_card.EditTagsFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					deck_ui.BuildFullFlex(deck_card.DetailText)
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
					deck_ui.BuildHelp(deck_card.EditTagsFlex, deck_help.HelpLabels)
				}
				return event
			})

			deck_ui.BuildFullFlex(deck_card.EditTagsFlex)
		} else if event.Rune() == 63 {
			// ? -> deck_help menu
			deck_ui.BuildHelp(deck_card.DetailText, deck_help.HelpView)
		}
		return event
	})

	deck_card.DetailEditText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			deck_card.DetailText.Clear()
			deck_card.DetailText.SetTitle(fmt.Sprintf(" %s ", deck_card.EditableCard.Title))
			deck_card.DetailText.SetText(deck_markdown.GetMarkDownDescription(utils.FormatDescription(deck_card.EditableCard.Description), configuration))
			deck_ui.BuildFullFlex(deck_card.DetailText)
		} else if event.Key() == tcell.KeyF2 {
			deck_card.EditableCard.Description = deck_card.DetailEditText.GetText()
			go deck_card.EditCard()
			deck_card.CardsMap[deck_card.EditableCard.Id] = deck_card.EditableCard
			deck_card.DetailText.SetText(deck_markdown.GetMarkDownDescription(utils.FormatDescription(deck_card.EditableCard.Description), configuration))
			deck_ui.BuildFullFlex(deck_card.DetailText)
		}
		return event
	})
	deck_card.DetailText.SetBorder(true)
	deck_card.DetailText.SetBorderColor(utils.GetColor(configuration.Color))

	deck_card.DetailEditText.SetBorder(true)
	deck_card.DetailEditText.SetBorderColor(utils.GetColor(configuration.Color))

	pages.AddPage("Main", deck_ui.FullFlex, true, true)

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func updateStacks() {
	for i, s := range deck_stack.Stacks {
		if s.Id == deck_card.EditableCard.StackId {
			for j, c := range s.Cards {
				if c.Id == deck_card.EditableCard.Id {
					deck_stack.Stacks[i].Cards[j] = deck_card.EditableCard
					break
				}
			}
			break
		}
	}
}
