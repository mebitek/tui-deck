package deck_comment

import (
	"fmt"
	"github.com/rivo/tview"
	"sort"
	"tui-deck/deck_http"
	"tui-deck/deck_markdown"
	"tui-deck/deck_structs"
	"tui-deck/deck_ui"
	"tui-deck/utils"
)

var Comments []deck_structs.Comment

var CommentTree *tview.TreeView
var CommentMap = make(map[int]deck_structs.Comment)

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

func getComments(cardId int) {

	var err error
	Comments, err = deck_http.GetComments(cardId, configuration)
	if err != nil {
		deck_ui.FooterBar.SetText(fmt.Sprintf("Error getting comments from card: %s", err.Error()))
	}

	replies := make(map[int]deck_structs.Comment)
	for _, c := range Comments {
		CommentMap[c.Id] = c
		cs := CommentStruct{
			Comment: c,
		}
		if c.ReplyTo != nil {
			replies[c.ReplyTo.Id] = c
		} else {
			CommentTreeStructMap[c.Id] = cs
		}
	}

	keySlice := make([]int, 0)
	for key, _ := range replies {
		keySlice = append(keySlice, key)
	}
	sort.Ints(keySlice)

	for _, k := range keySlice {
		r := replies[k]
		cs, ok := CommentTreeStructMap[k]
		if ok {
			c := CommentStruct{Comment: r}
			cs.Replies = append(cs.Replies, c)
			CommentTreeStructMap[k] = cs
		} else {
			for k1, r1 := range CommentTreeStructMap {
				searchReplies(r1.Replies, k, r)
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

	getComments(cardId)
	root := tview.NewTreeNode("COMMENTS").
		SetColor(utils.GetColor(configuration.Color)).SetSelectable(false)
	CommentTree.SetRoot(root).SetCurrentNode(root)

	keySlice := make([]int, 0)
	for key, _ := range CommentTreeStructMap {
		keySlice = append(keySlice, key)
	}
	sort.Ints(keySlice)
	for _, key := range keySlice {
		l := CommentTreeStructMap[key]
		comment := CommentMap[key]
		node := tview.NewTreeNode(fmt.Sprintf("#%d - %s", comment.Id, deck_markdown.GetMarkDownDescription(comment.Message, configuration)))
		buildTree(l.Replies, node)
		root.AddChild(node)
	}
}

func buildTree(replies []CommentStruct, node *tview.TreeNode) {
	if len(replies) > 0 {
		for _, r := range replies {
			node1 := tview.NewTreeNode(fmt.Sprintf("#%d - %s", r.Comment.Id, deck_markdown.GetMarkDownDescription(r.Comment.Message, configuration)))
			node.AddChild(node1)
			buildTree(r.Replies, node1)
		}
	}
}
