package auth

import "log"

var repository Repository

func GetUserByID(id string) (*User, error) {
	user, err := repository.Get(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetAllUsers() ([]*User, error) {
	users, err := repository.List()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser creates a user using the provided Email, Password, and DisplayName.
func CreateUser(user *CreateUserRequest) (*User, error) {
	createdUser, err := repository.Create(user)
	if err != nil {
		return nil, EmailExistsError
	}

	return createdUser, nil
}

func UpdateUser(user *UpdateUserRequest) (*User, error) {
	return &User{
		Profile:            nil,
		ID:                 "",
		Disabled:           false,
		CreationTimestamp:  0,
		LastLogInTimestamp: 0,
	}, nil
}

func DeleteUser(user *User) (*User, error) {
	err := repository.Delete(user.ID)
	return user, err
}

func init() {
	repo, err := NewFirebaseRepository()
	if err != nil {
		log.Panicf("error creating Firebase user repository: %v\n", err)
	}

	repository = repo
}
