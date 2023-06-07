package deck_comment

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sort"
	"strings"
	"time"
	"tui-deck/deck_http"
	"tui-deck/deck_markdown"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var Comments []deck_structs.Comment

var CommentsMap = make(map[int]deck_structs.Comment)
var CommentTree *tview.TreeView
var app *tview.Application
var Modal *tview.Modal
var configuration utils.Configuration

var CommentTreeStructMap = make(map[int]CommentStruct)

type CommentStruct struct {
	Comment deck_structs.Comment
	Replies []CommentStruct
}

func Init(application *tview.Application, conf utils.Configuration) {
	app = application
	configuration = conf

	CommentTree = tview.NewTreeView()
	CommentTree.SetBorder(true)
	CommentTree.SetBorderColor(utils.GetColor(configuration.Color))

	Modal = tview.NewModal()
}

func GetComments(cardId int) {

	CommentTreeStructMap = make(map[int]CommentStruct)

	var err error
	Comments, err = deck_http.GetComments(cardId, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting comments from card: %s", err.Error()))
	}

	replies := make(map[int][]deck_structs.Comment)
	for _, c := range Comments {
		CommentsMap[c.Id] = c
		cs := CommentStruct{
			Comment: c,
		}
		if c.ReplyTo != nil {
			replies[c.ReplyTo.Id] = append(replies[c.ReplyTo.Id], c)
		} else {
			CommentTreeStructMap[c.Id] = cs
		}
	}

	keySlice := make([]int, 0)
	for key := range replies {
		keySlice = append(keySlice, key)
	}
	sort.Ints(keySlice)

	for _, k := range keySlice {
		r := replies[k]
		cs, ok := CommentTreeStructMap[k]
		if ok {
			for _, r1 := range r {
				c := CommentStruct{Comment: r1}
				cs.Replies = append(cs.Replies, c)
			}
			CommentTreeStructMap[k] = cs
		} else {
			for k1, r1 := range CommentTreeStructMap {
				for _, r2 := range r {
					searchReplies(r1.Replies, k, r2)
				}
				CommentTreeStructMap[k1] = r1
			}
		}
	}
}

func searchReplies(replies []CommentStruct, k int, r deck_structs.Comment) {
	for i, r2 := range replies {
		if k == r2.Comment.Id {
			c := CommentStruct{Comment: r}
			r2.Replies = append(r2.Replies, c)
			replies[i] = r2
		} else {
			searchReplies(r2.Replies, k, r)
		}
	}
}

func CreateCommentsTree(cardId int) {
	GetComments(cardId)

	root := tview.NewTreeNode("COMMENTS").
		SetColor(utils.GetColor(configuration.Color)).SetSelectable(false)
	CommentTree.SetRoot(root).SetCurrentNode(root)

	keySlice := make([]int, 0)
	for key := range CommentTreeStructMap {
		keySlice = append(keySlice, key)
	}
	sort.Ints(keySlice)
	for _, key := range keySlice {
		l := CommentTreeStructMap[key]
		comment := l.Comment
		node := tview.NewTreeNode(fmt.Sprintf("[%s:-:-]#%d[-:-:-] - [%s:-:i][%s][-:-:-] - %s", configuration.Color,
			comment.Id, configuration.Color, getCreationDate(comment),
			deck_markdown.GetMarkDownDescription(comment.Message, configuration)))
		node.SetReference(l.Comment.Id)
		buildTree(l.Replies, node)
		root.AddChild(node)
	}
}

func buildTree(replies []CommentStruct, node *tview.TreeNode) {
	if len(replies) > 0 {
		for _, r := range replies {
			node1 := tview.NewTreeNode(fmt.Sprintf("#%d - [-:-:i][%s][-:-:-] - %s", r.Comment.Id,
				getCreationDate(r.Comment), deck_markdown.GetMarkDownDescription(r.Comment.Message, configuration)))
			node1.SetReference(r.Comment.Id)
			node.AddChild(node1)
			buildTree(r.Replies, node1)
		}
	}
}

func getCreationDate(comment deck_structs.Comment) string {
	parse, _ := time.Parse("2006-01-02T15:04:05+00:00", comment.CreationDateTime)
	return parse.Format("15:04:05 - 2006-01-02")
}

func BuildAddForm(c deck_structs.Comment) (*tview.Form, *deck_structs.Comment) {
	addForm := tview.NewForm()
	var comment = deck_structs.Comment{}
	var title = " Add Comment "
	if c.Id != 0 {
		comment = c
		title = " Edit Comment"
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
			deck_ui.BuildFullFlex(CommentTree)
			return nil
		}
		return event
	})
	addForm.AddTextArea("Message", c.Message, 60, 10, 300, func(message string) {
		comment.Message = message
	})

	return addForm, &comment
}

func AddComment(cardId int, comment deck_structs.Comment) {
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"message":"%s" }`, comment.Message), "\n", `\n`)
	var newComment deck_structs.Comment
	var err error
	newComment, err = deck_http.AddComment(cardId, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error creating new comment: %s", err.Error()))
	}

	CommentTreeStructMap[newComment.Id] = CommentStruct{Comment: newComment}
}

func EditComment(cardId int, comment deck_structs.Comment) {
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"message":"%s" }`, comment.Message), "\n", `\n`)
	_, err := deck_http.EditComment(cardId, comment.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error editing new comment: %s", err.Error()))
	}
	//_, ok := CommentTreeStructMap[comment.Id]
	//if ok {
	//	CommentTreeStructMap[comment.Id] = CommentStruct{Comment: comment, Replies: CommentTreeStructMap[comment.Id].Replies}
	//} else {
	//	for k, c := range CommentTreeStructMap {
	//		parentId, _ := findParent(c.Replies, editComment.Id)
	//		searchEditReplies(c.Replies, k, parentId, editComment.Id)
	//	}
	//}
}

func searchEditReplies(replies []CommentStruct, k int, parentId int, commentId int) {
	for i, r := range replies {
		if r.Comment.Id == parentId {
			for i1, r1 := range r.Replies {
				if r1.Comment.Id == commentId {
					r.Replies[i1] = CommentStruct{Comment: r.Comment, Replies: r.Replies[i1].Replies}
					CommentTreeStructMap[k] = CommentStruct{Comment: CommentTreeStructMap[k].Comment, Replies: r.Replies}
					break
				}
			}
		}
		if i == len(replies)-1 {
			searchEditReplies(r.Replies, k, parentId, commentId)
		}
	}
}

func ReplyComment(cardId int, parentId int, comment deck_structs.Comment) {
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"message":"%s", "parentId": %d }`, comment.Message, parentId), "\n", `\n`)
	//var newComment deck_structs.Comment
	var err error
	_, err = deck_http.AddComment(cardId, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error replying comment: %s", err.Error()))
	}

	/*	parentComment, ok := CommentTreeStructMap[parentId]
		if ok {
			rep := append(parentComment.Replies, CommentStruct{Comment: newComment})
			parentComment = CommentStruct{Comment: parentComment.Comment, Replies: rep}
			CommentTreeStructMap[parentId] = parentComment
		} else {
			for k1, r1 := range CommentTreeStructMap {
				searchReplies(r1.Replies, parentId, newComment)
				CommentTreeStructMap[k1] = r1
			}
		}*/

}

func DeleteComment(cardId int, commentId int) {

	Modal.ClearButtons()
	Modal.SetText(fmt.Sprintf("Are you sure to delete card #%d?", cardId))
	Modal.SetBackgroundColor(utils.GetColor(configuration.Color))

	Modal.AddButtons([]string{"Yes", "No"})

	Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			deck_ui.MainFlex.RemoveItem(Modal)
			app.SetFocus(CommentTree)
		}
		if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEnter {
			return event
		}
		return nil
	})

	Modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			go func() {
				_, err := deck_http.DeleteComment(cardId, commentId, configuration)
				if err != nil {
					deck_ui.FooterBar.SetText(fmt.Sprintf("Error deleting comment: %s", err.Error()))
				}
			}()

			CreateCommentsTree(cardId)
			deck_ui.FullFlex.RemoveItem(Modal)
			app.SetFocus(CommentTree)
		} else if buttonLabel == "No" {
			deck_ui.FullFlex.RemoveItem(Modal)
			app.SetFocus(CommentTree)
		}
	})

	deck_ui.FullFlex.AddItem(Modal, 0, 0, false)
	app.SetFocus(Modal)
}

func findParent(replies []CommentStruct, commentId int) (int, error) {
	var found = false
	var parentId int
	var listReplies = make([]CommentStruct, 0)
	for _, r := range replies {
		if r.Comment.Id == commentId {
			parentId = r.Comment.ReplyTo.Id
			found = true
		}
		for _, r1 := range r.Replies {
			listReplies = append(listReplies, r1)
		}

	}
	if !found {
		return findParent(listReplies, commentId)
	}
	return parentId, nil
}

func deleteReplies(replies []CommentStruct, k int, parentId int, commentId int) {
	var listReplies = make([]CommentStruct, 0)
	var found = false
	for i, r := range replies {
		if r.Comment.Id == parentId {
			for i1, r1 := range r.Replies {
				if r1.Comment.Id == commentId {
					r.Replies = append(r.Replies[:i1], r.Replies[i1+1:]...)
					replies[i].Replies = r.Replies
					found = true
					break
				}
			}
		} else {
			if r.Comment.Id == commentId {
				p := CommentTreeStructMap[parentId]
				p.Replies = append(p.Replies[:i], p.Replies[i+1:]...)
				CommentTreeStructMap[parentId] = CommentStruct{Comment: p.Comment, Replies: p.Replies}
				found = true
				break
			}
		}
		for _, r1 := range r.Replies {
			listReplies = append(listReplies, r1)
		}
	}
	if !found {
		deleteReplies(listReplies, k, parentId, commentId)
	}
}
