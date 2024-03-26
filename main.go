package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type User struct {
	Id       int
	Name     string
	Password string
}

func createUser(name string, password string) User {
	return User{
		Name:     name,
		Password: password,
	}
}

func (u *User) saveUser() (int64, error) {
	result, err := db.Exec("INSERT INTO user (username, password) VALUE (?, ?)", u.Name, u.Password)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, err
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, err := template.ParseGlob("view/*.html")
	if err != nil {
		return fmt.Errorf("error parsing template : %v", http.StatusInternalServerError)
	}

	err = tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		return fmt.Errorf("error parsing template: %q", http.StatusInternalServerError)
	}
	return nil
}

func loginValidation(name string, password string) bool {
	var isValid bool
	validUser := User{
		Name:     "shieldz",
		Password: "password",
	}

	if name != validUser.Name && password != validUser.Password {
		isValid = false
	} else {
		isValid = true
	}

	return isValid
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user := createUser(username, password)
	_, err := user.saveUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = renderTemplate(w, "register.html", user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	user := User{
		Name:     r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	isValid := loginValidation(user.Name, user.Password)

	if isValid {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err := renderTemplate(w, "login.html", user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {

	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "users",
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()

	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Database telah terkoneksi")

	http.HandleFunc("/", frontPageHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.ListenAndServe(":8080", nil)
}
