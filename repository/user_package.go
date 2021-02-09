package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// TODO Store in some sort of config
// TODO Relocate this file to a more sensical place
const STORAGE_PATH = "/home/leon/.luxbox"

type User struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

func (usr User) String() string {
	return fmt.Sprintf("%s %s", usr.ID, usr.Token)
}

// TODO Handle reading/writing in a more efficient way
type UserRepository struct {
	users map[string]User
}

func NewUserRepository() (*UserRepository, error) {
	path := fmt.Sprintf("%s/users.json", STORAGE_PATH)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("User list not found, (re)creating")
		err := ioutil.WriteFile(path, []byte("[]"), 0644)

		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, _ := ioutil.ReadAll(file)
	var users []User

	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	usrMap := make(map[string]User, len(users))
	for _, usr := range users {
		usrMap[usr.ID] = usr
	}

	return &UserRepository{usrMap}, nil
}

func (repo *UserRepository) All() map[string]User {
	return repo.users
}

func (repo *UserRepository) Find(id string) (User, error) {
	if usr, ok := repo.users[id]; ok {
		return usr, nil
	}

	return User{}, errors.New("user not found in repository")
}

func (repo *UserRepository) Save(usr *User) {
	repo.users[usr.ID] = *usr
}

func (repo *UserRepository) Flush() error {
	// When set to higher than zero, an empty LocalPackage is added
	s := make([]User, 0)
	for _, pkg := range repo.users {
		s = append(s, pkg)
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/users.json", STORAGE_PATH)
	err = ioutil.WriteFile(path, data, 644)

	return err
}
