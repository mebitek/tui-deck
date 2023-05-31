package deck_stack

import (
	"errors"
	"github.com/rivo/tview"
	"strings"
	"tui-deck/deck_structs"
	"tui-deck/utils"
)

var Stacks []deck_structs.Stack
var app *tview.Application
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration) {

	app = application
	configuration = conf

}

func GetActualStack(actualList tview.List) (int, deck_structs.Stack, error) {
	for i, s := range Stacks {
		if s.Title == strings.TrimSpace(actualList.GetTitle()) {
			return i, s, nil
		}
	}
	return 0, deck_structs.Stack{}, errors.New("not found")
}
