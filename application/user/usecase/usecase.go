package usecase

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/H-b-IO-T-O-H/kts-backend/application/user"
	"github.com/apsdehal/go-logger"
)

type UserUseCase struct {
	iLog   *logger.Logger
	errLog *logger.Logger
	repos  user.RepositoryUser
}

func NewUserUseCase(iLog *logger.Logger, errLog *logger.Logger,
	repos user.RepositoryUser) *UserUseCase {
	return &UserUseCase{
		iLog:   iLog,
		errLog: errLog,
		repos:  repos,
	}
}

func (u *UserUseCase) Login(user models.UserLogin) (*models.User, common.Err) {
	return u.repos.Login(user)
}

func (u *UserUseCase) CreateUserTemplate(newUser models.User) common.Err {
	return u.repos.CreateUserTemplate(newUser)
}
