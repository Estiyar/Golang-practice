package usecase

import (
	"practice3go/internal/repository"
	"practice3go/pkg/modules"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(r repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: r}
}

func (u *UserUsecase) GetUsers() ([]modules.User, error) {
	return u.repo.GetUsers()
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	return u.repo.GetUserByID(id)
}

func (u *UserUsecase) CreateUser(user modules.User) (int, error) {
	return u.repo.CreateUser(user)
}

func (u *UserUsecase) UpdateUser(id int, user modules.User) error {
	return u.repo.UpdateUser(id, user)
}

func (u *UserUsecase) DeleteUser(id int) (int64, error) {
	return u.repo.DeleteUserByID(id)
}

func (u *UserUsecase) GetPaginatedUsers(page int, pageSize int, filters map[string]string, orderBy string) (modules.PaginatedResponse, error) {
	return u.repo.GetPaginatedUsers(page, pageSize, filters, orderBy)
}

func (u *UserUsecase) GetCommonFriends(userID int, otherUserID int) ([]modules.User, error) {
	return u.repo.GetCommonFriends(userID, otherUserID)
}