package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store Storage, fname ,lname ,password string)*Account{
	acc,err:= NewAccount(fname,lname,password)
	if err !=nil{
		log.Fatal(err)
	}
	if err:=store.CreateAccount(acc);err!=nil{
		log.Fatal(err)
	}
	return acc
}

func seedAccounts(store Storage){
	acc:=seedAccount(store,"Numeez","Baloch","78692")
	fmt.Println(acc.Number)
}



func main (){
	seed:=flag.Bool("seed",false,"seed the db")
	flag.Parse()

	store,err:=NewPostgresStore()
	if err!=nil{
		log.Fatal(err)
	}
	if err := store.Init();err!=nil{
		log.Fatal(err)
	}
	if *seed{
		fmt.Println("Seeding the database")
		seedAccounts(store)
	}
	
	sever:=NewApiServer(":3000",store)
	sever.Run()
}