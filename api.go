package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error
type ApiError struct {
	Error string `json:"error"`
}

type ApiServer struct {
	listenAddr string
	store Storage
}

func NewApiServer(listenAdrr string, store Storage) *ApiServer {
	return &ApiServer{
		listenAddr: listenAdrr,
		store: store,
	}
}
func (s *ApiServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleGetAccountById))
	router.HandleFunc("/transfer",makeHttpHandleFunc(s.handleTransfer))
	log.Println("Json Api server running on the port", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *ApiServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)

}
func (s *ApiServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts,err:=s.store.GetAccounts()
	if err!=nil{
		return err
	}
	WriteJson(w,http.StatusOK,accounts)
	return nil
}
func (s *ApiServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET"{
	id,err:=getID(r)
	if err!=nil{
		return fmt.Errorf("invalid id given : %s",err)
	}
	account,err:=s.store.GetAccountById(id)
	if err!=nil{
		return err
	}
	return WriteJson(w, http.StatusOK, account)
}
if r.Method == "DELETE"{
	return s.handleDeleteAccount(w,r)
}
return fmt.Errorf("invalid method %s",r.Method)

}
func (s *ApiServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest:=new(CreateAccountRequest)
	if err:= json.NewDecoder(r.Body).Decode(createAccountRequest);err!=nil{
		return err
	}
	account:= NewAccount(createAccountRequest.FirstName,createAccountRequest.LastName)
	if err:=s.store.CreateAccount(account);err!=nil{
		return err
	}
	return WriteJson(w,http.StatusOK,account)

}
func (s *ApiServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
		id,err:=getID(r)
		if err!=nil{
			return err
		}
		if err:=s.store.DeleteAccount(id);err!=nil{
			return err
		}
		return WriteJson(w,http.StatusOK,map[string]int{"deleted":id})
}
func (s *ApiServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transfer:= new(TransferRequest)
	if err:=json.NewDecoder(r.Body).Decode(transfer);err!=nil{
		return err
	}
	defer r.Body.Close()
	return WriteJson(w,http.StatusOK,transfer)

}

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle error here
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
func getID(r *http.Request) (int,error){
	id:=mux.Vars(r)["id"]
	Id,err:=strconv.Atoi(id)
	if err!=nil{
		return Id,fmt.Errorf("invalid id is given %s",id)
	}
	return Id,err
}
