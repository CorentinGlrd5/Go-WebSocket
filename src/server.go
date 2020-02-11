package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"

	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// User ...
type User struct {
	ID       int
	Username string
	Password string
	Email    string
}

func hash(password string, key string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password+key), bcrypt.DefaultCost)
	if err != nil {
		return err.Error()
	}
	return string(hash)
}

func compare(hash string, password string, key string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+key))
	if err != nil {
		return false
	}
	return true
}

func createToken(db *sql.DB, user User, expiration int64) (*string, error) {
	u, err := uuid.NewV4()
	token := u.String()
	date := time.Now().Add(time.Duration(expiration) * time.Minute)
	_, err = db.Exec("INSERT INTO tokens (id, user_id, expiration) VALUES (?, ?, ?)", token, user.ID, date)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func fromToken(db *sql.DB, token string) User {
	var user User
	db.QueryRow("SELECT users.id, users.username, users.password, users.email FROM users JOIN tokens ON users.id = tokens.user_id WHERE tokens.id = ?", token).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	return user
}

func deleteExpiredTokens(db *sql.DB) (*int64, error) {
	res, err := db.Exec("DELETE FROM tokens WHERE expiration < NOW()")
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &rows, nil
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("mysql", os.Getenv("connectBDD"))
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	db.Exec("DROP TABLE IF EXISTS users")
	db.Exec("DROP TABLE IF EXISTS tokens")
	db.Exec("CREATE TABLE users (id SERIAL,username VARCHAR(255), password VARCHAR(255), email VARCHAR(255), data JSON, PRIMARY KEY (id))")
	db.Exec("CREATE TABLE tokens (id VARCHAR(255), user_id INT, expiration DATETIME)")

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var body User
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&body)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			var user User
			err = db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", body.Username).Scan(&user.ID, &user.Username, &user.Password)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			bool := compare(user.Password, body.Password, os.Getenv("KEY"))
			if bool {
				token, err := createToken(db, user, 61)
				if err != nil {
					http.Error(w, "Something went wrong", http.StatusUnauthorized)
					return
				}
				rows, err := deleteExpiredTokens(db)
				if err != nil {
					fmt.Fprintf(w, "%d %s", &rows, err.Error())
					return
				}
				http.SetCookie(w, &http.Cookie{Name: "token", Value: *token, HttpOnly: true, Expires: time.Now().Add(61 * time.Minute)})
				fmt.Fprintf(w, "OK")
				return
			}
			http.Error(w, "Password doesn't match", http.StatusUnauthorized)
		}
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var body User
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&body)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			var username string
			err = db.QueryRow("SELECT username FROM users WHERE username = ?", body.Username).Scan(&username)
			if err == nil {
				http.Error(w, "Username is already taken", http.StatusUnauthorized)
				return
			}
			hash := hash(body.Password, os.Getenv("KEY"))
			_, err = db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", body.Username, hash, body.Email)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "OK")
		}
	})

	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)
			json := buf.String()
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "No valid token", http.StatusUnauthorized)
				return
			}
			user := fromToken(db, cookie.Value)
			_, err = db.Exec("UPDATE users SET data = ? WHERE id = ?", json, user.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, json)
		}
	})

	http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "No valid token", http.StatusUnauthorized)
				return
			}
			user := fromToken(db, cookie.Value)
			var json string
			err = db.QueryRow("SELECT data FROM users WHERE id = ?", user.ID).Scan(&json)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, json)
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "No valid token", http.StatusUnauthorized)
				return
			}
			_, err = db.Exec("DELETE FROM tokens WHERE id = ?", cookie.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "token", HttpOnly: true, Expires: time.Now()})
			fmt.Fprintf(w, "You're logged out")
		}
	})

	http.Handle("/", http.FileServer(http.Dir("public")))
	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 1337 and log any errors
	log.Println("http server started on :1337")
	http.ListenAndServe(":1337", nil)
}
