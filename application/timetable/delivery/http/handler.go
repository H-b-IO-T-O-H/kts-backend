package http

import (
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common/models"
	"github.com/H-b-IO-T-O-H/kts-backend/application/timetable"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TimetableHandler struct {
	TimetableUseCase timetable.UseCase
	SessionBuilder   common.SessionBuilder
}

type Resp struct {
	Timetable *models.Timetable `json:"timetable"`
}

type RespWeek struct {
	Week *models.Week `json:"week"`
}

func NewRest(router *gin.RouterGroup, useCase timetable.UseCase, sessionBuilder common.SessionBuilder, AuthRequired gin.HandlerFunc) *TimetableHandler {
	rest := &TimetableHandler{TimetableUseCase: useCase, SessionBuilder: sessionBuilder}
	rest.routes(router, AuthRequired)
	return rest
}

func (u *TimetableHandler) routes(router *gin.RouterGroup, AuthRequired gin.HandlerFunc) {
	router.Use(AuthRequired)
	{
		router.GET("/:group_name/:week_number", u.GetTimetableWeekHandler)
		router.POST("/create", u.CreateTimetableHandler)
		router.DELETE("/", u.DeleteWeekHandler)
	}
}

func (u *TimetableHandler) CreateTimetableHandler(ctx *gin.Context) {
	session := u.SessionBuilder.Build(ctx)

	if userRole := session.Get("user_role"); userRole != common.Admin && userRole != common.Methodist {
		ctx.JSON(http.StatusForbidden, common.RespErr{Message: common.ForbiddenErr})
	}

	var template models.Week
	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, common.RespErr{Message: common.EmptyFieldErr})
		return
	}
	err := u.TimetableUseCase.CreateUpdateWeek(template)
	if err != nil {
		ctx.JSON(err.StatusCode(), err)
		return
	}
	ctx.JSON(http.StatusCreated, nil)
}

func (u *TimetableHandler) GetTimetableWeekHandler(ctx *gin.Context) {
	var req struct {
		WeekNumber int    `uri:"week_number" binding:"required"`
		GroupName  string `uri:"group_name" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&req); err != nil || req.WeekNumber > 2 || req.WeekNumber <= 0 {
		ctx.JSON(http.StatusBadRequest, common.RespErr{Message: common.EmptyFieldErr})
		return
	}
	buf, err := u.TimetableUseCase.GetTimetableWeek(req.GroupName, req.WeekNumber)
	if err != nil {
		ctx.JSON(err.StatusCode(), err)
		return
	}
	ctx.JSON(http.StatusOK, RespWeek{buf})
}

func (u *TimetableHandler) DeleteWeekHandler(ctx *gin.Context) {
	var template models.DeleteEntity
	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, common.RespErr{Message: common.EmptyFieldErr})
		return
	}
	session := u.SessionBuilder.Build(ctx)
	if userRole := session.Get("user_role"); userRole != common.Admin && userRole != common.Methodist {
		ctx.JSON(http.StatusForbidden, common.RespErr{Message: common.ForbiddenErr})
		return
	}
	err := u.TimetableUseCase.DeleteTimetableWeek(template)
	if err != nil {
		ctx.JSON(err.StatusCode(), err)
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
