package deck_structs

type Owner struct {
	PrimaryKey  string `json:"primaryKey"`
	Uid         string `json:"uid"`
	DisplayName string `json:"displayName"`
}

type Board struct {
	Id        int     `json:"id"`
	Title     string  `json:"title"`
	Owner     Owner   `json:"owner"`
	Color     string  `json:"color"`
	Labels    []Label `json:"labels"`
	Etag      string  `json:"etag"`
	Updated   bool    `json:"-"`
	DeletedAt int     `json:"deletedAt"`
}

type Stack struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Order int    `json:"order"`
	Cards []Card `json:"cards"`
}

type Card struct {
	Id          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Labels      []Label `json:"labels"`
	StackId     int     `json:"stackId"`
	Order       int     `json:"order"`
	Type        string  `json:"type"`
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

type Ocs struct {
	Meta Meta      `json:"meta"`
	Data []Comment `json:"data"`
}
type OcsSingle struct {
	Meta Meta    `json:"meta"`
	Data Comment `json:"data"`
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
