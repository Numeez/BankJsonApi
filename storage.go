package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	GetAccountByNumber(int) (*Account,error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	godotenv.Load()
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) Init() error {
	return s.createaAccountTable()

}
func (s *PostgresStore) createaAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS Account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		encrypted_password varchar(200),
		balance serial,
		created_at timestamp
	);`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `INSERT INTO Account
	 (first_name,last_name,number,encrypted_password,balance,created_at)
	 VALUES 
	 ($1,$2,$3,$4,$5,$6)`
	resp, err := s.db.Query(query, 
		account.FirstName,
		account.LastName,
		account.Number,
		account.EncryptedPassword,
		account.Balance,
		account.CreatedAt)
	if err != nil {
		return err
	}
	fmt.Println("Rows affected : ", resp)

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {

	return nil
}
func (s *PostgresStore) DeleteAccount(id int) error {
	_,err:=s.db.Query(`DELETE FROM Account WHERE id =$1`,id)
	if err!=nil{
		return err
	}
	
	return nil
}
func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	rows,err:=s.db.Query(`SELECT * FROM Account WHERE id=$1`,id)
	if err!=nil{
		return nil,err
	}

	for rows.Next(){
		return scanIntoAccounts(rows)
	}

	return nil,fmt.Errorf("account %d is not found",id)
}
func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	var accounts []*Account
	rows, err := s.db.Query(`SELECT * from Account;`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		account,err:=scanIntoAccounts(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil

}
func (s *PostgresStore) GetAccountByNumber(number int) (*Account,error) {
	rows,err:= s.db.Query("SELECT * FROM Account WHERE number = $1",number)
	if err!=nil{
		return nil,err
	}
	for rows.Next(){
		return scanIntoAccounts(rows)
	}
	return nil,fmt.Errorf("account with number [%d] not found",number)
}

func scanIntoAccounts(rows *sql.Rows) (*Account,error){
	account := new(Account)
	err := rows.Scan(
		&account.Id,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)
		
	return account,err
}
