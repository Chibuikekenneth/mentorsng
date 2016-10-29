package api

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/securecookie"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var hometpl *template.Template

var Database *sql.DB

var Server = "https://localhost:8080"

var Format string

type Users struct {
	Users []User `json:"users"`
}
type User struct {
	ID       int    "json:id"
	Name     string "json:username"
	Password string "json:password"
	Salt     string "json:salt"
	Email    string "json:email"
	First    string "json:first"
	Last     string "json:last"
}

func GetFormat(r *http.Request) {
	Format = r.URL.Query()["format"][1]

}

func SetFormat(data interface{}) []byte {
	var apiOutput []byte
	if Format == "json" {
		output, err := json.Marshal(data)
		if err != nil {
			fmt.Println("unable to marshall json")
		}
		apiOutput = output
	} else if Format == "xml" {
		output, err := xml.Marshal(data)
		if err != nil {
			fmt.Println("unable to marshall xml")
		}
		apiOutput = output
	}
	return apiOutput
}

type CreateResponse struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"code"`
}

type ErrMsg struct {
	ErrCode    int
	StatusCode int
	Msg        string
}

func ErrorMessages(err int64) (int, int, string) {
	errorMessage := ""
	statusCode := 200
	errorCode := 0
	switch err {
	case 1062:
		errorMessage = http.StatusText(409)
		errorCode = 10
		statusCode = http.StatusConflict
	}
	return errorCode, statusCode, errorMessage
}

func StartServer() {
	db, err := sql.Open("mysql", "go@/test")
	if err != nil {
		fmt.Println("unable to connect to database")
	}
	Database = db
	db.Ping()

}
func Init() {
	Routes := mux.NewRouter()
	Routes.HandleFunc("/api/users", CreateUser).Methods("POST")
	Routes.HandleFunc("/api/user/{id}", GetUser)
	Routes.HandleFunc("/api/users", UsersInfo).Methods("OPTIONS")
	Routes.HandleFunc("/api/users", RetrieveUsers).Methods("GET")
	Routes.HandleFunc("/api.{format:json|xml|txt}/user", RetrieveUsers).Methods("GET")
	Routes.HandleFunc("/api/users/{id:[0-9]+}", GetUser).Methods("GET")
	Routes.HandleFunc("/api/users/{id:[0-9]+}", UsersUpdate).Methods("PUT")
	Routes.HandleFunc("/login", LoginHandler).Methods("POST")
	Routes.HandleFunc("/logout", LogoutHandler).Methods("POST")
	Routes.HandleFunc("/internal", InternalPageHandler).Methods("GET")
	http.Handle("/welcome/", http.StripPrefix("/welcome/", http.FileServer(http.Dir("public"))))

	http.Handle("/", Routes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
	log.Println("Listening now from server...")

}

func UsersInfo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}

//CreateUser Method is used
func CreateUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Access-Control-Allow-Origin", Server)

	log.Println("starting user creation")

	NewUser := User{}

	NewUser.Email = r.FormValue("email")
	NewUser.Password = r.FormValue("password")
	//NewUser.First = r.FormValue("firstname")
	//NewUser.Last = r.FormValue("lastname")

	//output, err := json.Marshal(NewUser)

	//fmt.Println(output)

	response := CreateResponse{}

	stmt, err := Database.Prepare("insert into users ( user_email,user_password) values(?,?);")
	if err != nil {
		fmt.Print(err.Error())
		fmt.Println("unable to prepare statement")
	}
	result, err := stmt.Exec(NewUser.Email, NewUser.Password)

	if err != nil {
		if err != nil {
			errorMessage, errorCode := dbErrorParse(err.Error())
			fmt.Println(errorMessage)
			error, httpCode, msg := ErrorMessages(errorCode)
			response.Error = msg
			response.ErrorCode = error
			http.Error(w, "Conflict", httpCode)
		}

	}
	fmt.Println(result)
	createoutput, _ := json.Marshal(&response)
	fmt.Fprintln(w, string(createoutput))

}

//GetUser is used to retrieve user
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Pragma", "no-cache")

	urlParams := mux.Vars(r)
	id := urlParams["id"]
	ReadUser := User{}

	err := Database.QueryRow("select * from users where userid=?", id).Scan(&ReadUser.ID, &ReadUser.Name, &ReadUser.Password, &ReadUser.Salt, &ReadUser.First,
		&ReadUser.Last, &ReadUser.Email)

	switch {
	case err == sql.ErrNoRows:
		fmt.Fprintf(w, "No such user found ")
	case err != nil:
		log.Print(err)
		fmt.Fprintf(w, "Error")
	default:
		output, _ := json.Marshal(&ReadUser)
		fmt.Fprintf(w, string(output))
		log.Print(string(output))

	}

}

//RetrieveUsers is a function
func RetrieveUsers(w http.ResponseWriter, r *http.Request) {

	log.Println("starting retrieval")
	fmt.Println(r)
	//GetFormat(r)
	start := 0
	limit := 10
	next := start + limit
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Link", "<http://localhost:2000/api/users?start="+string(next)+"; rel=\"next\"")

	rows, _ := Database.Query("select * from users LIMIT 10")
	Response := Users{}
	for rows.Next() {
		user := User{}
		rows.Scan(&user.ID, &user.Name, &user.First, &user.Password, &user.Salt, &user.Last, &user.Email)
		Response.Users = append(Response.Users, user)
	}
	output, _ := json.Marshal(Response)

	fmt.Println(Response)
	fmt.Fprintln(w, string(output))
}

func dbErrorParse(err string) (string, int64) {
	Parts := strings.Split(err, ":")
	errorMessage := Parts[1]
	Code := strings.Split(Parts[0], "Error ")
	errorCode, _ := strconv.ParseInt(Code[1], 10, 32)
	return errorMessage, errorCode
}

func UsersUpdate(w http.ResponseWriter, r *http.Request) {

	fmt.Println("calling the put method")

	response := CreateResponse{}

	params := mux.Vars(r)
	uid := params["id"]

	email := r.FormValue("email")

	var userCount int

	err := Database.QueryRow("SELECT Count(userid) from users where userid=?", uid).Scan(&userCount)
	if userCount == 0 {

		error, httpcode, msg := ErrorMessages(404)
		log.Println(error)
		log.Println(w, msg, httpcode)
		response.Error = msg
		response.ErrorCode = httpcode
		http.Error(w, msg, httpcode)
	} else if err != nil {
		fmt.Println("update did not work")
	} else {
		_, upr := Database.Exec("Update users set user_last =? where userid=?", email, uid)
		if upr != nil {
			_, errorCode := dbErrorParse(upr.Error())
			_, httpCode, msg := ErrorMessages(errorCode)
			fmt.Println(upr)

			response.Error = msg
			response.ErrorCode = httpCode
			http.Error(w, msg, httpCode)
		} else {
			response.Error = "success"
			response.ErrorCode = 0
			output := SetFormat(response)
			fmt.Fprintln(w, string(output))
		}
	}
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

func LoginHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("email")
	pass := request.FormValue("password")
	redirectTarget := "/welcome"
	if name != "" && pass != "" {
		log.Println("this is a real login attempt")
		// .. check credentials ..
		setSession(name, response)
		redirectTarget = "/internal"
	}
	http.Redirect(response, request, redirectTarget, 302)
}

func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

const internalPage = "<h1>Internal</h1><hr><small>User: %s</small><form method='none'action='/logout'><button type='submit'>Logout</button></form>"

func InternalPageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		fmt.Fprintf(response, internalPage, userName)

	} else {
		http.Redirect(response, request, "/", 302)
	}
}
