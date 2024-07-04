package main

import (

	"log"
)



func main (){
	store,err:=NewPostgresStore()
	if err!=nil{
		log.Fatal(err)
	}
	if err := store.Init();err!=nil{
		log.Fatal(err)
	}
	sever:=NewApiServer(":3000",store)
	sever.Run()
}