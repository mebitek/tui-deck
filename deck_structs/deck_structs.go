package deck_structs

type Owner struct {
	PrimaryKey  string
	Uid         string
	DisplayName string
}

type Board struct {
	Id    int
	Title string
	Owner Owner
	Color string
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
}

type Label struct {
	Id    int
	Title string
	Color string
}
