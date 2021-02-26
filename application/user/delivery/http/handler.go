package http

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/H-b-IO-T-O-H/kts-backend/application/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UserHandler struct {
	UserUseCase    user.UseCase
	SessionBuilder common.SessionBuilder
}

type Resp struct {
	User *models.User `json:"user"`
}

func NewRest(router *gin.RouterGroup, useCase user.UseCase, sessionBuilder common.SessionBuilder, AuthRequired gin.HandlerFunc) *UserHandler {
	rest := &UserHandler{UserUseCase: useCase, SessionBuilder: sessionBuilder}
	rest.routes(router, AuthRequired)
	return rest
}

func (u *UserHandler) routes(router *gin.RouterGroup, AuthRequired gin.HandlerFunc) {
	router.POST("/login", u.LoginHandler)
	router.POST("/create", u.CreateUserHandler)
	router.Use(AuthRequired)
	{
		router.GET("/me", u.GetCurrentUser)
		router.POST("/logout", u.LogoutHandler)
	}
}

func (u *UserHandler) LoginHandler(ctx *gin.Context) {
	var reqUser models.UserLogin
	if err := ctx.ShouldBindJSON(&reqUser); err != nil {
		ctx.JSON(http.StatusForbidden, common.RespErr{Message: common.EmptyFieldErr})
		return
	}
	if err := common.ReqValidation(&reqUser); err != nil {
		ctx.JSON(http.StatusBadRequest, common.RespErr{Message: err.Error()})
		return
	}
	u.Login(ctx, reqUser)
}

func (u *UserHandler) Login(ctx *gin.Context, reqUser models.UserLogin) {
	buf, err := u.UserUseCase.Login(reqUser)
	if err != nil {
		ctx.JSON(err.StatusCode(), err)
		return
	}
	session := u.SessionBuilder.Build(ctx)
	if !reqUser.ChekBox {
		session.Options(sessions.Options{Domain: "10-tka.pp.ua", // for postman
			MaxAge:   2 * 3600,
			Secure:   true,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteNoneMode})
	}
	session.Set(common.UserRole, buf.Role)
	session.Set(common.UserId, buf.ID.String())
	if err := session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, common.RespErr{Message: common.SessionErr})
		return
	}

	ctx.JSON(http.StatusOK, Resp{User: buf})
}

func (u *UserHandler) LogoutHandler(ctx *gin.Context) {
	session := u.SessionBuilder.Build(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	err := session.Save()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.RespErr{Message: common.SessionErr})
		return
	}
	ctx.JSON(http.StatusOK, nil)
}

func (u *UserHandler) CreateUserHandler(ctx *gin.Context) {
	var reqUser struct {
		Role         string `json:"role" binding:"required"`
		Password     string `json:"password" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Surname      string `json:"surname" binding:"required"`
		Patronymic   string `json:"patronymic"`
		Email        string `json:"email" binding:"required" valid:"email"`
		StudentGroup string `json:"group"`
	}
	if err := ctx.ShouldBindJSON(&reqUser); err != nil {
		ctx.JSON(http.StatusBadRequest, common.RespErr{Message: common.EmptyFieldErr})
		return
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(reqUser.Password), bcrypt.DefaultCost)

	if err := u.UserUseCase.CreateUserTemplate(models.User{
		Role:         reqUser.Role,
		Name:         reqUser.Name,
		Surname:      reqUser.Surname,
		Patronymic:   reqUser.Patronymic,
		Email:        reqUser.Email,
		PasswordHash: passwordHash,
		StudentGroup: reqUser.StudentGroup,
	}); err != nil {
		ctx.JSON(err.StatusCode(), err)
		return
	}
	ctx.JSON(http.StatusOK, nil)
}

func (u *UserHandler) GetCurrentUser(ctx *gin.Context) {
	session := u.SessionBuilder.Build(ctx)
	userID := session.Get(common.UserId)

	id, _ := uuid.Parse(userID.(string))
	userById, err := u.UserUseCase.GetUserById(id)
	if err != nil {
		ctx.JSON(err.StatusCode(), err)
		return
	}

	ctx.JSON(http.StatusOK, Resp{User: userById})
}
