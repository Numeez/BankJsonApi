package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo3MzY3MzkxLCJleHBpcmVzQXQiOjE1MDAwfQ.Hf614ICFmSiEZ7PTdJMhtYsUV9Xa-N6NGfedjCqrwTU

type apiFunc func(w http.ResponseWriter, r *http.Request) error
type ApiError struct {
	Error string `json:"error"`
}

type ApiServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAdrr string, store Storage) *ApiServer {
	return &ApiServer{
		listenAddr: listenAdrr,
		store:      store,
	}
}
func (s *ApiServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login",makeHttpHandleFunc(s.handleLogin))
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleGetAccountById), s.store))
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))
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
//1075540
func (s *ApiServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method!="POST"{
		return fmt.Errorf("method not allowed %s",r.Method)
	}
	var loginRequest LoginRequest
	err:= json.NewDecoder(r.Body).Decode(&loginRequest)
	if err!=nil{
		return err
	}
	acc,err:= s.store.GetAccountByNumber(int(loginRequest.Number))
	if err!=nil{
		return err
	}
	if !acc.ValidatePassword(loginRequest.Password){
		return fmt.Errorf("not authenticated")
	}
	token,err:=createJWT(acc)
	if err!=nil{
		return err
	}
	
	resp:= LoginResponse{
		Token: token,
		Number: int(acc.Number),
	}

	return WriteJson(w,http.StatusOK,resp)
}

func (s *ApiServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	WriteJson(w, http.StatusOK, accounts)
	return nil
}
func (s *ApiServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return fmt.Errorf("invalid id given : %s", err)
		}
		account, err := s.store.GetAccountById(id)
		if err != nil {
			return err
		}
		return WriteJson(w, http.StatusOK, account)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("invalid method %s", r.Method)

}
func (s *ApiServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}
	account ,err:= NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName,createAccountRequest.Password)
	if err!=nil{
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}
	return WriteJson(w, http.StatusOK, account)

}
func (s *ApiServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJson(w, http.StatusOK, map[string]int{"deleted": id})
}
func (s *ApiServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transfer := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transfer); err != nil {
		return err
	}
	defer r.Body.Close()
	return WriteJson(w, http.StatusOK, transfer)

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
func getID(r *http.Request) (int, error) {
	id := mux.Vars(r)["id"]
	Id, err := strconv.Atoi(id)
	if err != nil {
		return Id, fmt.Errorf("invalid id is given %s", id)
	}
	return Id, err
}

func permissionDenied(w http.ResponseWriter) {
	WriteJson(w, http.StatusForbidden, ApiError{Error: "Permission Denied"})
}

func withJWTAuth(handleFunc http.HandlerFunc, store Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling jwt middleware")
		tokenString := r.Header.Get("jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return

		}
		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}
		account, err := store.GetAccountById(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		if account.Number != int64(claims["accountNumber"].(float64)) {
			permissionDenied(w)
			return
		}
		handleFunc(w, r)

	}
}

func validateJWT(token string) (*jwt.Token, error) {
	godotenv.Load()
	secret := os.Getenv("SECRET_TOKEN")
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}
func createJWT(account *Account) (string, error) {
	godotenv.Load()
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}
	secret := os.Getenv("SECRET_TOKEN")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
