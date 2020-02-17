# Go webserver

Webserver is in site coded in go. You can register, connect and chat with your friends on the same url.

## Some notion

1.1 :

    To store a password securely, we need three methods :

    - The hash, it is a function which makes it possible to transform the password into an incomprehensible character string.
    - The salt, it allows you to add a random header at the start of the password before hashing to improve password security.
    - The pepper, you must put a single character string for all passwords and store it in the code. On the add to the password with its salt and on hash all this to have a complex password to decipher.

1.2 :

    Once the password is secure, we want to check if the user who wants to log in correctly enters his password. for that we will use the CompareHashAndPassword function of bcrypt.go.
    This function allows you to compare if the password that the user enters is the same as that before he created at the outset.

## Installation

Use the GO language. Download [go](https://golang.org/doc/install/).

Use [git](https://git-scm.com/downloads/).

## Usage

First clone the project

```git
git clone https://github.com/CorentinGlrd5/Go-WebSocket.git

```

Second, go in folder Go-WebSocket

```bash
cd /Go-WebSocket/

```

Third, change setting file .env (for mysql)

```go
connectBDD="user:password@tcp(IP:port)/go"
KEY="key"
```

Fourth, go this project

```go
go run ./src/server.go

```

finaly, open the window and have fun !

http://localhost:1337/

## Contributing

I was alone for this school project.
