package deck_card

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
	"tui-deck/deck_http"
	"tui-deck/deck_markdown"
	"tui-deck/deck_stack"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var DetailText *tview.TextView
var DetailEditText *tview.TextArea
var EditTagsFlex *tview.Flex
var Modal *tview.Modal

var CardsMap = make(map[int]deck_structs.Card)
var EditableCard = deck_structs.Card{}

var currentBoard deck_structs.Board

var app *tview.Application
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration, board deck_structs.Board) {

	app = application
	configuration = conf

	DetailText = tview.NewTextView()
	DetailEditText = tview.NewTextArea()
	EditTagsFlex = tview.NewFlex()

	Modal = tview.NewModal()
	currentBoard = board
}

func moveCardToStack(todoList tview.List, key tcell.Key) {
	i := todoList.GetCurrentItem()
	name, _ := todoList.GetItemText(i)
	cardId := utils.GetId(name)
	card := CardsMap[cardId]

	primitive := app.GetFocus()
	actualPrimitiveIndex := deck_ui.Primitives[primitive]

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
		if card.StackId == len(deck_stack.Stacks) {
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
	CardsMap[card.Id] = card
	destList := deck_ui.GetNextFocus(actualPrimitiveIndex + operator).(*tview.List)
	destList.InsertItem(0, fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, card.Id, card.Title), labels, rune(0), nil)
	todoList.RemoveItem(i)
}

func AddCard(actualList tview.List, card deck_structs.Card) {
	var stackIndex, stack, _ = deck_stack.GetActualStack(actualList)

	jsonBody := fmt.Sprintf(`{"title":"%s", "description": "%s", "type": "plain", "order": 0}`, card.Title, card.Description)
	var newCard deck_structs.Card
	var err error
	newCard, err = deck_http.AddCard(currentBoard.Id, stack.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new card: %s", err.Error()))
	}

	actualList.InsertItem(0, fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, newCard.Id, newCard.Title), "", rune(0), nil)
	CardsMap[newCard.Id] = newCard
	DetailText.Clear()
	EditableCard = newCard
	deck_stack.Stacks[stackIndex].Cards = append(deck_stack.Stacks[stackIndex].Cards[:1], deck_stack.Stacks[stackIndex].Cards[0:]...)
	deck_stack.Stacks[stackIndex].Cards[0] = newCard
	DetailText.SetTitle(fmt.Sprintf(" %s ", newCard.Title))
	DetailText.SetText(utils.FormatDescription(newCard.Description))
	deck_ui.BuildFullFlex(DetailText)
}

func EditCard() {
	description := strings.ReplaceAll(EditableCard.Description, "\"", "\\\"")
	title := strings.ReplaceAll(EditableCard.Title, "\"", "\\\"")

	jsonBody := strings.ReplaceAll(
		fmt.Sprintf(`{"description": "%s", "title": "%s", "type": "plain", "owner":"%s"}`,
			description, title, configuration.User), "\n", `\n`)
	var err error
	_, err = deck_http.UpdateCard(currentBoard.Id, EditableCard.StackId, EditableCard.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error updating card: %s", err.Error()))
	}
}

func updateCard(boardId, stackId int, cardId int, jsonBody string) {
	_, err := deck_http.UpdateCard(boardId, stackId, cardId, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error moving card: %s", err.Error()))
		return
	}
}

func AssignLabel(jsonBody string) {
	_, err := deck_http.AssignLabel(currentBoard.Id, EditableCard.StackId, EditableCard.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting tag from card: %s", err.Error()))
	}
}

func DeleteLabel(jsonBody string) {
	_, err := deck_http.DeleteLabel(currentBoard.Id, EditableCard.StackId, EditableCard.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting tag from card: %s", err.Error()))
	}
}

func BuildStacks() {
	deck_ui.MainFlex.Clear()
	for index, s := range deck_stack.Stacks {
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
			CardsMap[card.Id] = card
			todoList.AddItem(fmt.Sprintf("[%s]#%d[white] - %s ", configuration.Color, card.Id, card.Title), labels, rune(0), nil)
		}

		todoList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			cardId := utils.GetId(name)

			DetailText.SetTitle(fmt.Sprintf(" %s ", CardsMap[cardId].Title))
			DetailText.SetDynamicColors(true)

			description := utils.FormatDescription(CardsMap[cardId].Description)
			DetailText.SetText(deck_markdown.GetMarkDownDescription(description, configuration))
			EditableCard = CardsMap[cardId]
			deck_ui.BuildFullFlex(DetailText)
		})

		todoList.SetFocusFunc(func() {
			todoList.SetTitleColor(utils.GetColor(configuration.Color))
		})

		deck_ui.Primitives[todoList] = index
		deck_ui.PrimitivesIndexMap[index] = todoList

		deck_ui.MainFlex.AddItem(todoList, 0, 1, true)
		primitive := deck_ui.MainFlex.GetItem(0)
		app.SetFocus(primitive)
	}
}
