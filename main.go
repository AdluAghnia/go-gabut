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

// TODO : Fix this Validator (DONE)
func loginValidation(user User, db *sql.DB) (bool, error) {
	username := user.Name
	var storedPassword []byte

	err := db.QueryRow("SELECT password from User WHERE username = ?", username).Scan(&storedPassword)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword(storedPassword, []byte(user.Password))
	if err != nil {
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

// TODO : Fix This Handler
func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user := createUser(r.FormValue("username"), r.FormValue("password"))
	err := renderTemplate(w, "login.html", user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	isValid, err := loginValidation(user, db)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "username or password is invalid", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if isValid {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
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
