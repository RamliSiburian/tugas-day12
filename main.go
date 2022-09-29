package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"personal-web/middleware"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// membuat route
	route := mux.NewRouter()

	// memanggil package connection
	connection.DatabaseConnect()

	// Membuat route path folder public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// route path uploads
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	// membuat routing ke setiap halaman yang akan ditampilkan (html)
	route.HandleFunc("/", index).Methods("GET")                                         // index.html
	route.HandleFunc("/form-project", formAddProject).Methods("GET")                    // addproject.html
	route.HandleFunc("/add-project", middleware.UploadFile(addProject)).Methods("POST") // menyimpan data dari form project ke db
	route.HandleFunc("/detail-project/{id_project}", detailProject).Methods("GET")      // menampilkan halaman detail project
	route.HandleFunc("/form-editproject/{id_project}", formEditProject).Methods("GET")  // menampilkan form editproject.html
	route.HandleFunc("/edit-project/{id_project}", editProject).Methods("POST")         // Menjalankan fungsi edit
	route.HandleFunc("/delete-project/{id_project}", deleteProject).Methods("GET")      // menjalankan fungsi delete
	route.HandleFunc("/form-register", formRegister).Methods("GET")                     // menampilkan halaman register.html
	route.HandleFunc("/register", register).Methods("POST")                             // menjalankan fungsi register
	route.HandleFunc("/form-login", formLogin).Methods("GET")                           // menampilkan halaman login.html
	route.HandleFunc("/login", login).Methods("POST")                                   // menjalankan fungsi login
	route.HandleFunc("/logout", logout).Methods("GET")                                  // menjalankan funsi logout

	// menjalankan server (port opsional)
	fmt.Println("Server running on port 5050")
	http.ListenAndServe("localhost:5050", route)

}

// untuk membuat stuct yang mendefenisikan tipe data yang akan ditampilka (dto = data transformation object)
type Project struct {
	ID               int
	ProjectName      string
	StartDate        time.Time
	EndDate          time.Time
	Format_startdate string
	Format_enddate   string
	Description      string
	Image            string
	Technologies     string
	Duration         string
	Author           string
	IsLogin          bool
}
type SessionData struct {
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = SessionData{}

type User struct {
	ID       int
	UserName string
	Email    string
	Password string
}

// function index / home
func index(w http.ResponseWriter, r *http.Request) {
	// membuat header/type html
	w.Header().Set("Description-Type", "text/html; charset=utf-8")

	// memanggil index.html dari folder views
	indeksTemplate, err := template.ParseFiles("views/index.html")

	if err != nil {
		// []byte untuk memberitahu bahwa data yang dikirim adalah tipe string
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["UserName"].(string)
	}
	flashmessage := session.Flashes("message")
	var flashes []string
	if len(flashmessage) > 0 {
		sessions.Save(r, w)
		for _, flash1 := range flashmessage {
			flashes = append(flashes, flash1.(string))
		}
	}
	// alert / flash message
	Data.FlashData = strings.Join(flashes, " ")
	// mengambil data dari database
	queryData, _ := connection.Conn.Query(context.Background(), "SELECT id_project, project_name, start_date, end_date, description, image, tb_user.username FROM tb_project LEFT JOIN tb_user ON tb_project.author = tb_user.id_user ORDER BY id_project DESC ")

	// untuk menampung data dari database yang disiman di struct Project
	var resultData []Project
	for queryData.Next() {
		newData := Project{}
		// men scan data dari struck Project "urutan data pada scan harus sesuai dengan query data"
		err := queryData.Scan(&newData.ID, &newData.ProjectName, &newData.StartDate, &newData.EndDate, &newData.Description, &newData.Image, &newData.Author)

		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(newData.Image)
		// data tanggal bentuk string
		newData.Format_startdate = newData.StartDate.Format("2006-01-02")
		newData.Format_enddate = newData.EndDate.Format("2006-01-02")
		startDateFormat := newData.StartDate.Format("2006-01-02")
		endDateFormat := newData.EndDate.Format("2006-01-02")
		layout := "2006-01-02"
		// data tanggal betuk time/date
		startDateParse, _ := time.Parse(layout, startDateFormat)
		endDateParse, _ := time.Parse(layout, endDateFormat)

		hours := endDateParse.Sub(startDateParse).Hours()
		days := hours / 24
		weeks := math.Round(days / 7)
		months := math.Round(days / 30)
		years := math.Round(days / 365)

		var duration string

		if days >= 1 && days <= 6 {
			duration = strconv.Itoa(int(days)) + " days"
		} else if days >= 7 && days <= 29 {
			duration = strconv.Itoa(int(weeks)) + " weeks"
		} else if days >= 30 && days <= 364 {
			duration = strconv.Itoa(int(months)) + " months"
		} else if days >= 365 {
			duration = strconv.Itoa(int(years)) + " years"
		}

		newData.Duration = duration

		// mem push data ke resultData
		resultData = append(resultData, newData)
	}

	// menalpilkan data ke html
	data := map[string]interface{}{
		"Projects":    resultData,
		"DataSession": Data,
	}
	// menampilkan index.html
	indeksTemplate.Execute(w, data)
}

// function form add project
func formAddProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Description-Type", "text/html; charset=utf-8")

	formAdd, err := template.ParseFiles("views/addproject.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	formAdd.Execute(w, nil)
}

// funtion add project
func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	// mengambil / menangkap data yang di input dari form
	projectName := r.PostForm.Get("projectName")
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")
	description := r.PostForm.Get("description")

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	// membuat session
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	// mendapatkan author id
	author := session.Values["ID"].(int)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(project_name , start_date , end_date , description, image, author) VALUES ($1, $2, $3, $4, $5, $6) ", projectName, startDate, endDate, description, image, author)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

// funtion detail project
func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	detailProjectTemplate, err := template.ParseFiles("views/detailproject.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var DetailProject = Project{}

	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id_project, project_name, start_date, end_date, description, tb_user.username FROM tb_project LEFT JOIN tb_user ON tb_project.author = tb_user.id_user WHERE id_project = $1", id_project).Scan(&DetailProject.ID, &DetailProject.ProjectName, &DetailProject.StartDate, &DetailProject.EndDate, &DetailProject.Description, &DetailProject.Author)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	DetailProject.Format_startdate = DetailProject.StartDate.Format("2006-01-02")
	DetailProject.Format_enddate = DetailProject.EndDate.Format("2006-01-02")
	startDateFormat := DetailProject.StartDate.Format("2006-01-02")
	endDateFormat := DetailProject.EndDate.Format("2006-01-02")
	layout := "2006-01-02"
	startDateParse, _ := time.Parse(layout, startDateFormat)
	endDateParse, _ := time.Parse(layout, endDateFormat)

	hours := endDateParse.Sub(startDateParse).Hours()
	days := hours / 24
	weeks := math.Round(days / 7)
	months := math.Round(days / 30)
	years := math.Round(days / 365)

	var duration string

	if days >= 1 && days <= 6 {
		duration = strconv.Itoa(int(days)) + " days"
	} else if days >= 7 && days <= 29 {
		duration = strconv.Itoa(int(weeks)) + " weeks"
	} else if days >= 30 && days <= 364 {
		duration = strconv.Itoa(int(months)) + " months"
	} else if days >= 365 {
		duration = strconv.Itoa(int(years)) + " years"
	}

	DetailProject.Duration = duration

	data := map[string]interface{}{
		"Project": DetailProject,
	}

	detailProjectTemplate.Execute(w, data)
}

// form edit project
func formEditProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Description-Type", "text/html; charset=utf-8")
	formEditTemplate, err := template.ParseFiles("views/editproject.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" + err.Error()))
		return
	}

	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	EditProject := Project{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT id_project, project_name, start_date, end_date, description FROM tb_project WHERE id_project = $1", id_project).Scan(&EditProject.ID, &EditProject.ProjectName, &EditProject.StartDate, &EditProject.EndDate, &EditProject.Description)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	EditProject.Format_startdate = EditProject.StartDate.Format("2006-01-02")
	EditProject.Format_enddate = EditProject.EndDate.Format("2006-01-02")

	data := map[string]interface{}{
		"Edits": EditProject,
	}

	formEditTemplate.Execute(w, data)
}

// edit project
func editProject(w http.ResponseWriter, r *http.Request) {
	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	// mengambil / menangkap data yang di input dari form
	projectNameEdit := r.PostForm.Get("projectName")
	startDateEdit := r.PostForm.Get("startDate")
	endDateEdit := r.PostForm.Get("endDate")
	descriptionEdit := r.PostForm.Get("description")

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_project SET project_name = $1, start_date = $2, end_date = $3, description = $4 WHERE id_project = $5", projectNameEdit, startDateEdit, endDateEdit, descriptionEdit, id_project)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	fmt.Println(startDateEdit)
	fmt.Println(projectNameEdit)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

// funtion delete project
func deleteProject(w http.ResponseWriter, r *http.Request) {
	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id_project=$1", id_project)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// funtion form register

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Description-Type", "text/html; charset=utf-8")
	registerTemplate, err := template.ParseFiles("views/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["UserName"].(string)
	}
	data := map[string]interface{}{
		"DataSession": Data,
	}

	registerTemplate.Execute(w, data)
}

// function register
func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	username := r.PostForm.Get("username")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(username, email, password) VALUES ($1, $2, $3)", username, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

// function form login
func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Description-Type", "text/html; charset=utf-8")
	loginTemplate, err := template.ParseFiles("views/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	flashMessage := session.Flashes("message")

	var flashes []string
	if len(flashMessage) > 0 {
		session.Save(r, w)
		for _, f1 := range flashMessage {
			flashes = append(flashes, f1.(string))
		}
	}
	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["UserName"].(string)
	}
	Data.FlashData = strings.Join(flashes, "")
	data := map[string]interface{}{
		"DataSession": Data,
	}

	loginTemplate.Execute(w, data)

}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var email = r.PostForm.Get("email")
	var password = r.PostForm.Get("password")

	// menampung data user dan di simpan di struct user
	user := User{}

	// mengambil data email, dan melakukan pengecekan email
	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.UserName, &user.Email, &user.Password)

	// melakukan pengecekan email
	if err != nil {
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Email atau passwor tidak cocok!", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		return
	}

	// melakukan pencocokan password dari form dengan db
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	// melakukan pengecekan password
	if err != nil {
		// fmt.Println("Password salah")
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Email atau password tidak cocok!", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	// berfungsi untuk menyimpan data kedalam session browser
	session.Values["UserName"] = user.UserName
	session.Values["Email"] = user.Email
	session.Values["ID"] = user.ID
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 10800 // detik

	session.AddFlash("Succesfull login", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/form-login", http.StatusSeeOther)
}
