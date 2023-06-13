package deck_structs

import (
	"fmt"
	"strings"
)

type Owner struct {
	PrimaryKey  string `json:"primaryKey"`
	Uid         string `json:"uid"`
	DisplayName string `json:"displayName"`
}

func (owner *Owner) GetAbbrv() string {
	split := strings.Split(owner.DisplayName, " ")

	if len(split) > 1 {
		return strings.ToUpper(fmt.Sprintf("%c%c", split[0][0], split[1][0]))
	}
	return strings.ToUpper(fmt.Sprintf("%c", owner.DisplayName[0]))
}

type Board struct {
	Id             int     `json:"id"`
	Title          string  `json:"title"`
	Owner          Owner   `json:"owner"`
	Color          string  `json:"color"`
	Labels         []Label `json:"labels"`
	Etag           string  `json:"etag"`
	Updated        bool    `json:"-"`
	CreateDefaults bool    `json:"-"`
	DeletedAt      int     `json:"deletedAt"`
	Users          []Owner `json:"users"`
}

type Stack struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Order int    `json:"order"`
	Cards []Card `json:"cards"`
}

type Card struct {
	Id            int            `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Labels        []Label        `json:"labels"`
	StackId       int            `json:"stackId"`
	Order         int            `json:"order"`
	Type          string         `json:"type"`
	DueDate       string         `json:"duedate"`
	AssignedUsers []AssignedUser `json:"assignedUsers"`
}

type AssignedUser struct {
	Id          int   `json:"id"`
	CardId      int   `json:"cardId"`
	Type        int   `json:"type"`
	Participant Owner `json:"participant"`
}

type Label struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
}

type OcsResponse struct {
	Ocs Ocs `json:"ocs"`
}

type OcsResponseSingle struct {
	Ocs OcsSingle `json:"ocs"`
}

type OcsResponseUsers struct {
	Ocs OcsUsers `json:"ocs"`
}

type Ocs struct {
	Meta Meta      `json:"meta"`
	Data []Comment `json:"data"`
}
type OcsSingle struct {
	Meta Meta    `json:"meta"`
	Data Comment `json:"data"`
}

type OcsUsers struct {
	Meta Meta  `json:"meta"`
	Data Users `json:"data"`
}

type Users struct {
	Users []string
}

type Meta struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type Comment struct {
	Id               int       `json:"id"`
	ObjectId         int       `json:"objectId"`
	Message          string    `json:"message"`
	ActorId          string    `json:"actorId"`
	ActorType        string    `json:"actorType"`
	ActorDisplayName string    `json:"actorDisplayName"`
	CreationDateTime string    `json:"creationDateTime"`
	Mentions         []Mention `json:"mentions"`
	ReplyTo          *Comment  `json:"replyTo"`
}

type Mention struct {
	MentionId          int    `json:"mentionId"`
	MentionType        string `json:"mentionType"`
	MentionDisplayName string `json:"mentionDisplayName"`
}
