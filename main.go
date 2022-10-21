package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var store Store

func main() {

	fmt.Println("Starting server...")

	//connStr format should be for example: "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	//see https://godoc.org/github.com/lib/pq for more details
	connString := os.Getenv("PG_CONNSTRING")
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfuly connected to databse!")
	InitStore(&dbStore{db: db})

	r := newRouter()
	fmt.Println("Serving on port 8080")
	http.ListenAndServe(":8080", r)
}

func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func newRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/hello", helloServer).Methods("GET")
	r.HandleFunc("/presents", GetPresentHandler).Methods("GET")
	r.HandleFunc("/presents", CreatePresentHandler).Methods("POST")
	r.HandleFunc("/totalbudget", GetTotalBudgetHandler).Methods("GET")

	staticFileDirectory := http.Dir("./static/")

	staticFileHandler := http.StripPrefix("/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/").Handler(staticFileHandler).Methods("GET")
	return r
}

//Store functions to interact wtih DB
func (store *dbStore) CreatePresent(present *Present) error {
	_, err := store.db.Query("INSERT INTO whishlist(person, present, budget) VALUES ($1, $2, $3)", present.Person, present.Present, present.Budget)
	return err
}

func (store *dbStore) GetPresent() ([]*Present, error) {
	rows, err := store.db.Query("SELECT person, present, budget FROM whishlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	presents := []*Present{}

	for rows.Next() {
		rowWithPresent := &Present{}
		if err := rows.Scan(&rowWithPresent.Person, &rowWithPresent.Present, &rowWithPresent.Budget); err != nil {
			return nil, err
		}
		presents = append(presents, rowWithPresent)
	}
	return presents, nil
}

func (store *dbStore) GetTotalBudget() ([]*TotalBudget, error) {
	rows, err := store.db.Query("SELECT SUM (budget) AS Total from whishlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalBudget := []*TotalBudget{}

	for rows.Next() {
		rowWithBudget := &TotalBudget{}
		if err := rows.Scan(&rowWithBudget.Total); err != nil {
			return nil, err
		}
		totalBudget = append(totalBudget, rowWithBudget)
	}
	return totalBudget, nil
}

func InitStore(s Store) {
	store = s
}

type Store interface {
	CreatePresent(present *Present) error
	GetPresent() ([]*Present, error)
	GetTotalBudget() ([]*TotalBudget, error)
}

type dbStore struct {
	db *sql.DB
}

func GetPresentHandler(w http.ResponseWriter, r *http.Request) {
	presents, err := store.GetPresent()

	presentsListBytes, err := json.Marshal(presents)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(presentsListBytes)
}

func GetTotalBudgetHandler(w http.ResponseWriter, r *http.Request) {
	totalBudget, err := store.GetTotalBudget()

	totalBudgetListBytes, err := json.Marshal(totalBudget)

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(totalBudgetListBytes)
}

func CreatePresentHandler(w http.ResponseWriter, r *http.Request) {
	present := Present{}

	err := r.ParseForm()

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	present.Budget = r.Form.Get("budget")
	present.Person = r.Form.Get("person")
	present.Present = r.Form.Get("present")

	err = store.CreatePresent(&present)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

var presents []Present

type Present struct {
	Budget  string `json:"budget"`
	Person  string `json:"person"`
	Present string `json:"present"`
}

type TotalBudget struct {
	Total int64 `json:"total"`
}
