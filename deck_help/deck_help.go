package deck_help

import (
	"github.com/rivo/tview"
)

var HelpMain = tview.NewTextView()
var HelpEdit = tview.NewTextView()

func InitHelp() {
	HelpMain = getHelp()
	HelpEdit = getHelp2()
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
[yellow]q[white]: Quit app.
[yellow]?[white]: Help.

[blue]Press Enter for more help, press Escape to return.`)
	HelpMain.SetTitle(" HELP - Main ")
	return HelpMain
}

func getHelp2() *tview.TextView {
	HelpEdit = tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Editing[white]

[yellow]e[white]: Edit card Description.
[yellow]t[white]: Edit card tags.
[yellow]F2[white]: Save card.
[yellow]ESC[white]: Back to main view.

[blue]Press Enter for more help, press Escape to return.`)
	HelpEdit.SetTitle(" HELP - Editing ")
	return HelpEdit
}
