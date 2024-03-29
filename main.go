package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Name     string
	Password string
}

func hashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return hash, err
}

func intiliazeDB() (*sql.DB, error) {
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "users",
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	pingErr := db.Ping()

	if pingErr != nil {
		return nil, pingErr
	}

	return db, nil
}

func createUser(name string, password string) User {
	return User{
		Name:     name,
		Password: password,
	}
}

func (u *User) saveUser(db *sql.DB) (int64, error) {
	hash, err := hashPassword(u.Password)
	if err != nil {
		return 0, err
	}
	result, err := db.Exec("INSERT INTO User (username, password) VALUE (?, ?)", u.Name, hash)
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

// TODO : Fix this Validator
func loginValidation(user User, db *sql.DB) (bool, error) {
	username := user.Name
	Password, err := hashPassword(user.Password)
	if err != nil {
		return false, err
	}
	var storedPassword string

	err = db.QueryRow("SELECT password from User WHERE username = ?", username).Scan(&storedPassword)
	if err != nil {
		return false, err
	}

	if string(Password) != storedPassword {
		return false, nil
	}

	return true, nil
}

func registerHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user := createUser(username, password)
	id, err := user.saveUser(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	user.Id = int(id)

	err = renderTemplate(w, "register.html", user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user := User{
		Name:     r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	isValid, err := loginValidation(user, db)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "username or password is invalid", http.StatusUnauthorized)
		} else {
			log.Fatal(err)
		}
	}
	if !isValid {
		http.Error(w, "Unautherized", http.StatusUnauthorized)
		return
	}

	err = renderTemplate(w, "login.html", user)
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
	db, err := intiliazeDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", frontPageHandler)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, db)
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		registerHandler(w, r, db)
	})
	http.ListenAndServe(":8080", nil)
}
