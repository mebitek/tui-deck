package deck_help

import (
	"github.com/rivo/tview"
)

var HelpMain = tview.NewTextView()
var HelpView = tview.NewTextView()
var HelpEdit = tview.NewTextView()
var HelpLabels = tview.NewTextView()
var HelpBoards = tview.NewTextView()

func InitHelp() {
	HelpMain = getHelp()
	HelpView = getHelp2()
	HelpEdit = getHelp3()
	HelpLabels = getHelp4()
	HelpBoards = getHelp5()
}

func getHelp() *tview.TextView {
	HelpMain = tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Main[white]

[yellow]TAB[white]: Switch stack.
[yellow]Down arrow[white]: Move down.
[yellow]Up arrow[white]: Move up.
[yellow]Right arrow[white]: Move card to next stack.
[yellow]Left arrow[white]: Move card to previous stack.
[yellow]ENTER[white]: Select card.
[yellow]s[white]: Switch board.
[yellow]r[white]: Reload board.
[yellow]a[white]: Add card to current stack.
[yellow]d[white]: Delete selected card in current stack.
[yellow]ctrl+a[white]: Add stack.
[yellow]ctrl+d[white]: Delete selected current stack.
[yellow]ctrl+e[white]: Edit current stack.
[yellow]q[white]: Quit app.
[yellow]?[white]: Help.

[blue]Press Enter for more help, press Escape to return.`)
	HelpMain.SetTitle(" HELP - Main ")
	return HelpMain
}

func getHelp2() *tview.TextView {
	HelpView = tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]View Card[white]

[yellow]e[white]: Edit card Description.
[yellow]t[white]: Edit card tags.
[yellow]ESC[white]: Back to main view.

[blue]Press Enter for more help, press Escape to return.`)
	HelpView.SetTitle(" HELP - View Card ")
	return HelpView
}
func getHelp3() *tview.TextView {
	HelpEdit = tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Edit Card[white]

Type to enter text.
[yellow]F2[white]: Save card.
[yellow]ESC[white]: Back to card view.

[blue]Press Enter for more help, press Escape to return.`)
	HelpEdit.SetTitle(" HELP - Edit Card ")
	return HelpEdit
}

func getHelp4() *tview.TextView {
	HelpLabels = tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Edit Card Labels[white]

[yellow]Up arrow[white]: Move up.
[yellow]Down arrow[white]: Move down.
[yellow]TAB[white]: Switch between card labels and available labels lists.
[yellow]ENTER[white]: If car label has been selected, delete it. If available label has been selected, add it to card
[yellow]ESC[white]: Back to card view.

[blue]Press Enter for more help, press Escape to return.`)
	HelpLabels.SetTitle(" HELP - Edit Card Labels ")
	return HelpLabels
}
func getHelp5() *tview.TextView {
	HelpBoards = tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Switch Boards[white]

[yellow]Up arrow[white]: Move up.
[yellow]Down arrow[white]: Move down.
[yellow]ENTER[white]: Select board.
[yellow]a[white]: Add board.
[yellow]e[white]: Edit board.
[yellow]d[white]: Delete board.
[yellow]ESC[white]: Back to main view.

[blue]Press Enter for more help, press Escape to return.`)
	HelpBoards.SetTitle(" HELP - Switch Boards ")
	return HelpBoards
}
