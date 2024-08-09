package models

type User struct {
	Id       string `db:"id"`
	Email    string `db:"email"`
	UserType string
	PassHash []byte `db:"password_hash"`
}
