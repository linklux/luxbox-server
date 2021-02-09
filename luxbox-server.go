package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/linklux/luxbox-server/action"
	"github.com/linklux/luxbox-server/repository"
)

// TODO Store in some sort of config
const STORAGE_PATH = "/home/leon/.luxbox"

var actions = map[string]interface{ action.IAction }{
	"register": action.RegisterAction{},
}

func handle(conn net.Conn) {
	fmt.Printf("handing connection from: %s\n", conn.RemoteAddr().String())
	defer conn.Close()

	req, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		conn.Write([]byte("ERR: failed to read input: " + err.Error() + "\n"))
		return
	}

	req = strings.TrimSpace(string(req))

	params := make(map[string]string)
	args := strings.Split(req, ";")

	for _, arg := range args {
		data := strings.Split(arg, ":")

		if len(data) != 2 {
			conn.Write([]byte("ERR: parameters must be in key:value format\n"))
			continue
		}

		params[data[0]] = data[1]
	}

	userRepository, err := repository.NewUserRepository()
	if err != nil {
		conn.Write([]byte("ERR: server error: " + err.Error() + "\n"))
	}

	userParam := ""
	if user, ok := params["user"]; ok {
		userParam = user
	}

	user, userErr := userRepository.Find(userParam)

	// TODO Handle user authentication
	if _, ok := params["action"]; ok {
		if _, ok := actions[params["action"]]; !ok {
			conn.Write([]byte("ERR: action '" + params["action"] + "' not supported\n"))
			return
		}

		handler := actions[params["action"]].New()

		if handler.RequireUserAuth() {
			if userErr != nil {
				conn.Write([]byte("ERR: user not found\n"))
				return
			}

			if token, ok := params["token"]; !ok || user.Token != token {
				conn.Write([]byte("ERR: user authentication failed\n"))
				return
			}
		}

		delete(params, "action")
		request := action.Request{User: &user, Conn: conn, Params: params}

		if err := handler.Validate(&request); err == nil {
			response := handler.Handle(&request)

			conn.Write([]byte(fmt.Sprintf("code:%d;payload:%s", response.Code, response.Payload)))
		} else {
			conn.Write([]byte("ERR: invalid request payload: " + err.Error() + "\n"))
		}
	} else {
		conn.Write([]byte("ERR: missing action in request payload\n"))
	}
}

func main() {
	fmt.Println("starting...")

	if _, err := os.Stat(STORAGE_PATH); os.IsNotExist(err) {
		err := os.MkdirAll(STORAGE_PATH, os.ModePerm)

		if err != nil {
			fmt.Printf("ERR: %s\n1", err.Error())
			return
		}
	}

	// listen on port 8068
	l, _ := net.Listen("tcp", ":8068")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handle(c)
	}
}
