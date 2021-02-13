package action

import (
	"fmt"

	"github.com/linklux/luxbox-server/component"
	"github.com/linklux/luxbox-server/repository"
)

// TODO Don't make this a client requestable action
type RegisterAction struct {
	component.StringGenerator
}

func (this RegisterAction) RequireUserAuth() bool { return false }

func (this RegisterAction) New() IAction {
	return RegisterAction{}
}

func (this RegisterAction) Validate(request *Request) error {
	return nil
}

func (this RegisterAction) Handle(request *Request) Response {
	userRepository, err := repository.NewUserRepository()
	if err != nil {
		return Response{-1, map[string]interface{}{
			"error": "server error: " + err.Error(),
		}}
	}

	uuid := this.GenerateUuid4()
	token, err := this.GenerateString(32)
	if err != nil {
		return Response{-1, map[string]interface{}{
			"error": "failed to generate token",
		}}
	}

	user := repository.User{ID: uuid, Token: token}

	userRepository.Save(&user)
	userRepository.Flush()

	fmt.Printf("registered user %s\n", uuid)

	return Response{0, map[string]interface{}{
		"id":    uuid,
		"token": token,
	}}
}
