package repository

import "github.com/pact-foundation/pact-workshop-go/model"

// UserRepository is an in-memory db representation of our set of users
type UserRepository struct {
	Users map[string]*model.User
}

// GetUsers returns all users in the repository
func (u *UserRepository) GetUsers() []model.User {
	var response []model.User

	for _, user := range u.Users {
		response = append(response, *user)
	}

	return response
}

// ByUsername finds a user by their username.
func (u *UserRepository) ByUsername(username string) (*model.User, error) {
	if user, ok := u.Users[username]; ok {
		return user, nil
	}
	return nil, model.ErrNotFound
}

// ByID finds a user by their ID
func (u *UserRepository) ByID(ID int) (*model.User, error) {
	for _, user := range u.Users {
		if user.ID == ID {
			return user, nil
		}
	}
	return nil, model.ErrNotFound
}
