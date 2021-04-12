package data

type User struct {
	ID       string `db:"user_id"`
	Token    string `db:"user_token"`
	Username string `db:"user_name"`
}

type Book struct {
	Name string
	Rows int
}

type Question struct {
	ID       string `db:"question_id"`
	Question string `db:"question_data"`
	Answer   string `db:"question_answer"`
	BData    string `db:"question_book"`
	Owner    string `db:"question_owner"`
}

type FromAskData struct {
	Name string
	Row  int
}
