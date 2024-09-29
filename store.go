package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	dbName string
}

func NewStore(config Config) *Store {
	cache := &Store{dbName: config.DbName}
	cache.CreateDb()
	return cache
}

func (c *Store) CreateDb() {
	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	createAccountTable := `
		CREATE TABLE IF NOT EXISTS account (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			password TEXT NOT NULL
   		);
	`

	tryPanic(db, createAccountTable)
}

func tryPanic(db *sql.DB, sqlStatement string) {
	_, err := db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
}

type AccountCredentials struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *Store) GetAccount(email string) AccountCredentials {
	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, email, password FROM account WHERE email = ?", email)

	var account AccountCredentials
	err = row.Scan(&account.Id, &account.Email, &account.Password)
	if err != nil {
		panic(err)
	}

	return account
}

func (c *Store) GetMailboxes(accountId int) []string {
	return []string{"INBOX"}
}
