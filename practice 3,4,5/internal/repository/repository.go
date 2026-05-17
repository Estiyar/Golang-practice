package repository

import (
	"practice3go/internal/repository/_postgres"
	"practice3go/internal/repository/_postgres/users"
	"practice3go/pkg/modules"
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(u modules.User) (int, error)
	UpdateUser(id int, u modules.User) error
	DeleteUserByID(id int) (int64, error)

	GetPaginatedUsers(page int, pageSize int, filters map[string]string, orderBy string) (modules.PaginatedResponse, error)
	GetCommonFriends(userID int, otherUserID int) ([]modules.User, error)
}

type Repositories struct {
	UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}