package deck_structs

type Owner struct {
	PrimaryKey  string
	Uid         string
	DisplayName string
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
	Id    int
	Title string
	Order int
	Cards []Card
}

type Card struct {
	Id          int
	Title       string
	Description string
	Labels      []Label
	StackId     int
	Order       int
	Type        string
}

type Label struct {
	Id    int
	Title string
	Color string
}
