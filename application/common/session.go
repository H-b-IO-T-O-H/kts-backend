package common

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SessionBuilder interface {
	Build(ctx *gin.Context) sessions.Session
}

type NewSessionBuilder struct {}

func (sb *NewSessionBuilder) Build (ctx *gin.Context) sessions.Session {
	return sessions.Default(ctx)
}

/*
func GetCurrentUserId(session sessions.Session, userType string) (id uuid.UUID, err error) {
	userIDStr := session.Get(userType)
	if userIDStr == nil {
		return uuid.Nil, nil
	}
	userID, err := uuid.Parse(userIDStr.(string))

	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

*/
