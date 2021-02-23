package common

import "strings"

const (
	EmptyFieldErr   = "обязательные поля не заполнены или содержат недопустимые данные"
	SessionErr      = "ошибка авторизации. Попробуйте авторизоваться повторно"
	ForbiddenErr    = "недостаточно прав"
	AuthRequiredErr = "необходима авторизация"
	ServerErr       = "что-то пошло не так. Попробуйте позже"
	UserExistErr    = "пользователь уже существует"
	AuthErr         = "пользователь с такими данными не зарегистрирован"
	NotFound        = "по данному запросу ничего не нашлось"
	WrongPasswd     = "неверное имя пользователя или пароль"
)

const (
	Admin     = "admin"
	Guest     = "guest"
	Methodist = "methodist"
	Student   = "student"
	Professor = "professor"
)

type Err interface {
	Msg() string
	StatusCode() int
}

type RespErr struct {
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (r RespErr) Msg() string {
	return r.Message
}

func (r RespErr) StatusCode() int {
	return r.Status
}

func NewErr(statusCode int, message string) Err {
	return RespErr{Status: statusCode, Message: message}
}

func RecordExists(errMsg string) bool {
	return strings.Contains(errMsg, "duplicate")
}

func NoRows(errMsg string) bool {
	return strings.Contains(errMsg, "no rows")
}
