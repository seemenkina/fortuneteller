package models

type User struct {
	Username string
	Token    string
}

type BookData struct {
	Name string
	Row  int
}

type Question struct {
	Question string
	Answer   string
	BData    BookData
	Owner    User
}
