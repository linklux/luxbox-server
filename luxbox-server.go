package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/linklux/luxbox-server/action"
	"github.com/linklux/luxbox-server/repository"
)

// TODO Store in some sort of config
const STORAGE_PATH = "/home/leon/.luxbox"

type requestData struct {
	Action string                 `json:"action"`
	Meta   map[string]interface{} `json:"meta"`
}

type responseData struct {
	Code int                    `json:"code"`
	Data map[string]interface{} `json:"data"`
}

var actions = map[string]interface{ action.IAction }{
	"register": action.RegisterAction{},
	"upload":   action.UploadAction{},
}

func response(conn net.Conn, code int, data map[string]interface{}) {
	response, err := json.Marshal(responseData{code, data})

	if err == nil {
		conn.Write(append(response, '\n'))
	} else {
		fmt.Printf("ERR: failed to send response: %s\n", err.Error())
	}
}

func error(conn net.Conn, code int, msg string) {
	response(conn, code, map[string]interface{}{"error": msg})
}

func stringHashB64(val string) string {
	hash := sha256.New()
	hash.Write([]byte(val))

	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func handle(conn net.Conn) {
	fmt.Printf("handing connection from: %s\n", conn.RemoteAddr().String())
	defer conn.Close()

	// Read the header information
	req, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		error(conn, -1, "failed to read input: "+err.Error())
		return
	}

	// Remove whitespace from start/end of header, including newline characters
	req = strings.TrimSpace(string(req))

	data := requestData{}
	if err = json.Unmarshal([]byte(req), &data); err != nil {
		error(conn, -1, "could not decode request: "+err.Error())
		return
	}

	// Create user repository and attempt to find the given user, if provided
	userRepository, err := repository.NewUserRepository()
	if err != nil {
		error(conn, -1, "server error:"+err.Error())
	}

	userParam := ""
	if user, ok := data.Meta["user"]; ok {
		userParam = user.(string)
	}

	user, userErr := userRepository.Find(userParam)

	// Find the configured helper for the given action
	if _, ok := actions[data.Action]; !ok {
		error(conn, -1, "action '"+data.Action+"' not supported")
		return
	}

	handler := actions[data.Action].New()

	// If the given handler requires user authentication, enforce it
	if handler.RequireUserAuth() {
		if userErr != nil {
			error(conn, -1, "user not found")
			return
		}

		if _, ok := data.Meta["token"]; !ok {
			error(conn, -1, "token parameter missing in request meta")
			return
		}

		storedToken := stringHashB64(user.Token)

		if data.Meta["token"] != string(storedToken) {
			error(conn, -1, "user authentication failed")
			return
		}
	}

	request := action.Request{User: &user, Conn: conn, Meta: data.Meta}

	// Validate the request with the handler, and handle if successful
	if err := handler.Validate(&request); err == nil {
		res := handler.Handle(&request)
		response(conn, res.Code, res.Data)
	} else {
		error(conn, -1, "invalid request payload: "+err.Error())
	}
}

func main() {
	fmt.Println("starting...")

	// Create the storage path when it does not exist yet
	if _, err := os.Stat(STORAGE_PATH); os.IsNotExist(err) {
		err := os.MkdirAll(STORAGE_PATH, os.ModePerm)

		if err != nil {
			fmt.Printf("ERR: %s\n1", err.Error())
			return
		}
	}

	// Listen on port 8068
	l, _ := net.Listen("tcp", ":8068")

	for {
		// Attempt to accept all incoming connections
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		// Connection OK, dispatch to goroutine and continue listening to new connections
		go handle(c)
	}
}
