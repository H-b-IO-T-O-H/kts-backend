package user

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/google/uuid"
)

type UseCase interface {
	Login(user models.UserLogin) (*models.User, common.Err)
	CreateUserTemplate(newUser models.User) common.Err
	GetUserById(userId uuid.UUID) (*models.User, common.Err)
}
