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

var CommentTreeStructMap = make(map[int]*CommentStruct)

type CommentStruct struct {
	Comment deck_structs.Comment
	Parent  *CommentStruct
	Replies []*CommentStruct
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

	CommentTreeStructMap = make(map[int]*CommentStruct)

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
			CommentTreeStructMap[c.Id] = &cs
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
				c := CommentStruct{Comment: r1, Parent: cs}
				cs.Replies = append(cs.Replies, &c)
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

func searchReplies(replies []*CommentStruct, k int, r deck_structs.Comment) {
	for i, r2 := range replies {
		if k == r2.Comment.Id {
			c := CommentStruct{Comment: r, Parent: r2}
			r2.Replies = append(r2.Replies, &c)
			replies[i] = r2
		} else {
			searchReplies(r2.Replies, k, r)
		}
	}
}

func CreateCommentsTree() {
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
		node := tview.NewTreeNode(fmt.Sprintf("[%s:-:-]#%d[-:-:-] - [%s:-:i]%s - [%s][-:-:-] - %s", configuration.Color,
			comment.Id, configuration.Color, comment.ActorDisplayName, getCreationDate(comment),
			deck_markdown.GetMarkDownDescription(comment.Message, configuration)))
		node.SetReference(l.Comment.Id)
		buildTree(l.Replies, node)
		root.AddChild(node)
	}
}

func buildTree(replies []*CommentStruct, node *tview.TreeNode) {
	if len(replies) > 0 {
		for _, r := range replies {
			node1 := tview.NewTreeNode(fmt.Sprintf("#%d - [-:-:i]%s - [%s][-:-:-] - %s", r.Comment.Id, r.Comment.ActorDisplayName,
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
	CommentTreeStructMap[newComment.Id] = &CommentStruct{Comment: newComment}
}

func EditComment(cardId int, comment deck_structs.Comment) {
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"message":"%s" }`, comment.Message), "\n", `\n`)
	editComment, err := deck_http.EditComment(cardId, comment.Id, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error editing new comment: %s", err.Error()))
	} else {
		for _, k := range CommentTreeStructMap {
			node := findById(k, comment.Id)
			if node != nil {
				node.Comment = editComment
				break
			}
		}
	}
}

func ReplyComment(cardId int, parentId int, comment deck_structs.Comment) {
	jsonBody := strings.ReplaceAll(fmt.Sprintf(`{"message":"%s", "parentId": %d }`, comment.Message, parentId), "\n", `\n`)
	//var newComment deck_structs.Comment
	newComment, err := deck_http.AddComment(cardId, jsonBody, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error replying comment: %s", err.Error()))
	} else {
		for _, k := range CommentTreeStructMap {
			node := findById(k, parentId)
			if node != nil {
				node.addReply(newComment)
				break
			}
		}
	}
}

func DeleteComment(cardId int, commentId int) {

	Modal.ClearButtons()
	Modal.SetText(fmt.Sprintf("Are you sure to delete comment #%d?", cardId))
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

			for _, k := range CommentTreeStructMap {
				node := findById(k, commentId)
				if node != nil {
					node.remove()
					break
				}
			}
			CreateCommentsTree()
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

func findById(root *CommentStruct, id int) *CommentStruct {
	queue := make([]*CommentStruct, 0)
	queue = append(queue, root)
	for len(queue) > 0 {
		nextUp := queue[0]
		queue = queue[1:]
		if nextUp.Comment.Id == id {
			return nextUp
		}
		if len(nextUp.Replies) > 0 {
			for _, child := range nextUp.Replies {
				queue = append(queue, child)
			}
		}
	}
	return nil
}

func (node *CommentStruct) remove() {
	if node.Parent != nil {
		for idx, sibling := range node.Parent.Replies {
			if sibling == node {
				node.Parent.Replies = append(
					node.Parent.Replies[:idx],
					node.Parent.Replies[idx+1:]...,
				)
			}
		}
	} else {
		delete(CommentTreeStructMap, node.Comment.Id)
	}

	if len(node.Replies) != 0 {
		for _, child := range node.Replies {
			child.Parent = nil
		}
		node.Replies = nil
	}
}

func (node *CommentStruct) addReply(newComment deck_structs.Comment) {
	node.Replies = append(node.Replies, &CommentStruct{Comment: newComment, Parent: node})
}
