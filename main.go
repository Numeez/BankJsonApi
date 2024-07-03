package main



func main (){
	sever:=NewApiServer(":3000")
	sever.Run()
}