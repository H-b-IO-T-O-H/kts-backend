package repository

import (
	"fmt"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/H-b-IO-T-O-H/kts-backend/application/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
)

type pgStorage struct {
	db *gorm.DB
}

func NewPgRepository(db *gorm.DB) user.RepositoryUser {
	return &pgStorage{db: db}
}

func (p *pgStorage) Login(user models.UserLogin) (*models.User, common.Err) {
	userDB := new(models.User)

	err := p.db.Take(userDB, "email = ?", user.Email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.RespErr{Message: common.AuthErr, Status: http.StatusNotFound}
		}
		return nil, common.RespErr{Message: err.Error(), Status: http.StatusInternalServerError}
	}
	// compare password with the hashed one
	err = bcrypt.CompareHashAndPassword(userDB.PasswordHash, []byte(user.Password))
	if err != nil {
		return nil, common.RespErr{Message: common.AuthErr, Status: http.StatusNotFound}
	}
	if userDB.Role == common.Student {
		err = p.db.Raw("select g.group_name from public.groups g join students s on s.group_id=g.group_id where s.user_id=?", userDB.ID).Row().Scan(&userDB.StudentGroup)
		if err != nil {
			return nil, common.RespErr{Message: common.AuthErr, Status: http.StatusNotFound}
		}
	}
	return userDB, nil
}

func (p *pgStorage) CreateUserTemplate(newUser models.User) common.Err {
	if err := p.db.Create(&newUser).Error; err != nil {
		msg := err.Error()
		if common.RecordExists(msg) {
			return common.RespErr{Message: common.UserExistErr, Status: http.StatusConflict}
		}
		return common.RespErr{Message: err.Error(), Status: http.StatusInternalServerError}
	}

	if newUser.Role == "student" {
		group := new(models.Group)
		group.GroupName = newUser.StudentGroup
		err := p.db.Take(&group).Where("group_name = ?", group.GroupName).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if p.db.Create(group).Error != nil {
					return common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
				}
			} else {
				return common.RespErr{Status: http.StatusInternalServerError, Message: err.Error()}
			}
		}
		err = p.db.Exec(fmt.Sprintf("insert into public.students(group_id, user_id) values ('%s', '%s')", group.GroupId, newUser.ID)).Error
	}
	return nil
}

func (p *pgStorage) GetUserById(userId uuid.UUID) (*models.User, common.Err) {
	userDB := new(models.User)

	userDB.ID = userId
	if err := p.db.Take(userDB).Error; err != nil {
		msg := err.Error()
		if common.NoRows(msg) {
			return nil, common.RespErr{Message: common.AuthErr, Status: http.StatusNotFound}
		} else {
			return nil, common.RespErr{Status: http.StatusInternalServerError, Message: msg}
		}
	}
	return userDB, nil
}