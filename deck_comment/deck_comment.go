package deck_comment

import (
	"fmt"
	"github.com/rivo/tview"
	"time"
	"tui-deck/deck_http"
	"tui-deck/deck_markdown"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var Comments []deck_structs.Comment

var CurrentCard deck_structs.Card
var CommentsList *tview.List
var CommentMap = make(map[int]deck_structs.Comment)

var DetailCommentView *tview.TextView

var app *tview.Application
var Modal *tview.Modal
var configuration utils.Configuration

func Init(application *tview.Application, conf utils.Configuration) {
	app = application
	configuration = conf
	CommentsList = tview.NewList()
	CommentsList.SetBorder(true)
	CommentsList.SetBorderColor(utils.GetColor(configuration.Color))

	DetailCommentView = tview.NewTextView()
	DetailCommentView.SetDynamicColors(true)
	DetailCommentView.SetBorder(true)
	DetailCommentView.SetBorderColor(utils.GetColor(configuration.Color))

	Modal = tview.NewModal()
}

func GetComments(cardId int) {

	var err error
	Comments, err = deck_http.GetComments(cardId, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting comments from card: %s", err.Error()))
	}

	CommentsList.Clear()
	for _, c := range Comments {
		secondaryMessage := ""
		if c.ReplyTo != nil {
			secondaryMessage = fmt.Sprintf("Reply to comment #%d", c.ReplyTo.Id)
		}
		parse, _ := time.Parse("2006-01-02T15:04:05+00:00", c.CreationDateTime)
		creationDate := parse.Format("15:04:05 - 2006-01-02")
		CommentsList.AddItem(fmt.Sprintf("#%d - %s", c.Id, creationDate), secondaryMessage, rune(0), nil)
		CommentMap[c.Id] = c
	}

	CommentsList.SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
		commentId := utils.GetId(name)
		DetailCommentView.SetTitle(fmt.Sprintf(" #%d - %s ", commentId, "Comment"))

		text := deck_markdown.GetMarkDownDescription(utils.FormatDescription(Comments[index].Message), configuration)
		comment := CommentMap[commentId]

		if comment.ReplyTo != nil {
			text = fmt.Sprintf("%s\n#%d -------------------------\n\n[%s:-:-]%s[-:-:-]", text, commentId, configuration.Color, deck_markdown.GetMarkDownDescription(comment.ReplyTo.Message, configuration))
		}

		DetailCommentView.SetText(text)
		deck_ui.BuildFullFlex(DetailCommentView)

	})

}
