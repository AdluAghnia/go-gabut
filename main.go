package main

import (
	"database/sql"
	"errors"
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
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "gobut",
		AllowNativePasswords: true,
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

func (u *User) validateRegister(db *sql.DB) (bool, error) {
	// CHECK IF PASSWORD HAVE MORE THAN 6 characters
	valid := len(u.Name) >= 3 && len(u.Password) >= 6
	// Check if username already exist in the database
	statment, err := db.Prepare("SELECT COUNT(*) FROM User WHERE username = ? ")

	if err != nil {
		return false, err
	}
	defer statment.Close()

	var count int
	err = statment.QueryRow(u.Name).Scan(&count)
	if err != nil {
		return false, err
	}

	if count != 0 {
		return false, errors.New("username already exist")
	}

	if !valid {
		return false, errors.New("check your username and password again")
	}

	if valid && count == 0 {
		return true, nil
	}

	return false, nil
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

func renderTemplate(w http.ResponseWriter, name string, err_message string) error {
	tmpl, err := template.ParseGlob("view/*.html")
	if err != nil {
		return fmt.Errorf("error parsing template : %v", err)
	}

	err = tmpl.ExecuteTemplate(w, name, err_message)
	if err != nil {
		return fmt.Errorf("error parsing template: %q", err)
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
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		user := createUser(username, password)

		isValid, err := user.validateRegister(db)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}

		if isValid {
			id, err := user.saveUser(db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("CREATING USER FOR ID %d SUCCES", id)

			http.Redirect(w, r, "/login", http.StatusSeeOther)

		}
	}
	err := renderTemplate(w, "register.html", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// TODO : Fix This Handler
func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		user := createUser(r.FormValue("username"), r.FormValue("password"))
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
	err := renderTemplate(w, "login.html", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "index.html", "")
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
