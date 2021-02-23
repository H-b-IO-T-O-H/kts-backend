package user

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
)

type RepositoryUser interface {
	Login(user models.UserLogin) (*models.User, common.Err)
	CreateUserTemplate(newUser models.User) common.Err
}