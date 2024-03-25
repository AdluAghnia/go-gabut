package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type User struct {
	Name     string
	Password string
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

func loginHandler(w http.ResponseWriter, r *http.Request) {

	user := User{
		Name:     r.FormValue("username"),
		Password: r.FormValue("password"),
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
	http.HandleFunc("/", frontPageHandler)
	http.HandleFunc("/login", loginHandler)
	http.ListenAndServe(":8080", nil)
}
