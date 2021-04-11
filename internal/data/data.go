package data

type User struct {
	ID       string `db:"user_id"`
	Token    string `db:"user_token"`
	Username string `db:"user_name"`
}

type Book struct {
	ID   string   `db:"book_id"`
	Name string   `db:"book_name"`
	Rows int      `db:"book_len"`
	Data []string `db:"book_data"`
}

type Question struct {
	ID       string `db:"question_id"`
	Question string `db:"question_data"`
	Answer   string `db:"question_answer"`
	BData    string `db:"question_book"`
	Owner    string `db:"question_owner"`
}

type BookData struct {
	Name string
	Row  int
}
