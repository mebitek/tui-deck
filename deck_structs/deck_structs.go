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
