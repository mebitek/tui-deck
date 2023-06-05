package deck_stack

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"tui-deck/deck_http"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var Stacks []deck_structs.Stack
var app *tview.Application
var Modal *tview.Modal
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration) {

	app = application
	configuration = conf
	Modal = tview.NewModal()
}
func GetActualStack(actualList *tview.List) (int, deck_structs.Stack, error) {
	for i, s := range Stacks {
		if s.Title == strings.TrimSpace(actualList.GetTitle()) {
			return i, s, nil
		}
	}
	return 0, deck_structs.Stack{}, errors.New("not found")
}

func AddStack(boardId int, stack deck_structs.Stack) {
	jsonBody := fmt.Sprintf(`{"title":"%s", "order": %d}`, stack.Title, stack.Order)
	var newStack deck_structs.Stack
	var err error
	newStack, err = deck_http.AddStack(boardId, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error crating new stack: %s", err.Error()))
	}

	Stacks = append(Stacks, newStack)

}

func DeleteStack(stackId int, actualList *tview.List) {
	Modal.ClearButtons()
	Modal.SetText(fmt.Sprintf("Are you sure to delete stack #%d?", stackId))
	Modal.SetBackgroundColor(utils.GetColor(configuration.Color))

	Modal.AddButtons([]string{"Yes", "No"})

	Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			deck_ui.MainFlex.RemoveItem(Modal)
			app.SetFocus(actualList)
		}
		if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEnter {
			return event
		}
		return nil
	})

	deck_ui.MainFlex.AddItem(Modal, 0, 0, false)
	app.SetFocus(Modal)
}

func EditStack(boardId int, stack deck_structs.Stack) {
	description := strings.ReplaceAll(stack.Title, "\"", "\\\"")

	jsonBody := strings.ReplaceAll(
		fmt.Sprintf(`{"title": "%s", "order": %d }`,
			description, stack.Order), "\n", `\n`)
	var err error
	_, err = deck_http.EditStack(boardId, stack.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error updating stack: %s", err.Error()))
	}
}

func BuildAddForm(s deck_structs.Stack) (*tview.Form, *deck_structs.Stack) {
	addForm := tview.NewForm()
	var stack = deck_structs.Stack{}
	var title = " Add Stack "
	if s.Id != 0 {
		stack = s
		title = " Edit Stack"
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
			deck_ui.BuildFullFlex(deck_ui.MainFlex)
			return nil
		}
		return event
	})
	addForm.AddInputField("Title", s.Title, 20, nil, func(title string) {
		stack.Title = title
	})
	addForm.AddInputField("Order", strconv.Itoa(s.Order), 5, func(textToCheck string, lastChar rune) bool {
		if lastChar < 48 || lastChar > 57 {
			return false
		}
		return true
	}, func(order string) {
		orderInt, _ := strconv.Atoi(order)
		stack.Order = orderInt
	})

	return addForm, &stack
}
