package action

import (
	"net"

	"github.com/linklux/luxbox-server/repository"
)

type IAction interface {
	New() IAction

	RequireUserAuth() bool

	Validate(request *Request) error
	Handle(request *Request) Response
}

type Request struct {
	User *repository.User
	Conn net.Conn

	Params map[string]string
}

type Response struct {
	Code    int
	Payload string
}
