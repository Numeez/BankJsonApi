package main

import (
	"math/rand"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct{
	Number int64 `json:"number"`
	Password string `json:"password"`
}

type TransferRequest struct{
	ToAccount int `json:"toAccount"`
	Amount int `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Password string `json:"password"`
}

type LoginResponse struct{
	Number int `json:"number`
	Token string `json:"token"`
}

type Account  struct {
	Id int `json:"id"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Number int64 `json:"number"`
	EncryptedPassword string `json:"-"`
	Balance int64 `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

func (a *Account)ValidatePassword(password string)bool{
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword),[]byte(password))==nil

}


func NewAccount (firstName,lastName, password string )(*Account,error){
	encpw,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err!=nil{
		return nil,err
	}
	return &Account{
		FirstName: firstName,
		LastName: lastName,
		Number: int64(rand.Intn(10000000)),
		EncryptedPassword: string(encpw),
		CreatedAt: time.Now().UTC(),
	},nil
}