package application

import (
	"context"
	"fmt"
	"github.com/H-b-IO-T-O-H/kts-backend/application/common"
	SessionBuilder "github.com/H-b-IO-T-O-H/kts-backend/application/common"
	TimetableHandler "github.com/H-b-IO-T-O-H/kts-backend/application/timetable/delivery/http"
	TimetableRepository "github.com/H-b-IO-T-O-H/kts-backend/application/timetable/repository"
	TimetableUseCase "github.com/H-b-IO-T-O-H/kts-backend/application/timetable/usecase"
	UserHandler "github.com/H-b-IO-T-O-H/kts-backend/application/user/delivery/http"
	UserRepository "github.com/H-b-IO-T-O-H/kts-backend/application/user/repository"
	UserUseCase "github.com/H-b-IO-T-O-H/kts-backend/application/user/usecase"
	"github.com/apsdehal/go-logger"
	"github.com/asaskevich/govalidator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type AuthCookieConfig struct {
	Key      string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

type Logger struct {
	Info  *logger.Logger
	Error *logger.Logger
}

type Config struct {
	Listen  string   `yaml:"listen"`
	Db      DBConfig `yaml:"db"`
	DocPath string   `yaml:"docPath"`
	Redis   string   `yaml:"redis_address"`
}

type App struct {
	config   Config
	log      Logger
	doneChan chan bool
	route    *gin.Engine
	db       *gorm.DB
}

func NewApp(config Config) *App {
	gin.Default()
	r := gin.New()

	infoLogger, _ := logger.New("Info logger", 1, os.Stdout)
	errorLogger, _ := logger.New("Error logger", 2, os.Stderr)
	infoLogger.SetLogLevel(logger.DebugLevel)

	log := Logger{
		Info:  infoLogger,
		Error: errorLogger,
	}

	credentials := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d", config.Db.User,
		config.Db.Password, config.Db.Name,
		config.Db.Host, config.Db.Port)
	db, err := gorm.Open(postgres.Open(credentials), &gorm.Config{})
	if err != nil {
		log.Error.Fatal("connection to postgres db failed...")
	}
	r.Use(common.RequestLogger(log.Info))
	r.Use(common.ErrorLogger(log.Error))
	r.Use(common.ErrorMiddleware())
	r.Use(common.Recovery())
	r.Use(common.Cors())
	r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})
	r.GET("/health", healthCheck())

	store, err := redis.NewStore(10, "tcp", config.Redis, "", []byte("secret"))
	if err != nil {
		log.Error.Fatal("connection to redis db failed...")
		os.Exit(-1)
	}

	store.Options(sessions.Options{
		//Domain:   "localhost", // for postman
		Domain: "135.181.207.76",
		MaxAge:   int((3 * 12 * time.Hour).Seconds()),
		Secure:   false,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		//SameSite: http.SameSiteStrictMode, // prevent csrf attack
	})
	govalidator.SetFieldsRequiredByDefault(false)

	sessionsMiddleware := sessions.Sessions("timetable", store)
	r.Use(sessionsMiddleware)
	sessionBuilder := SessionBuilder.NewSessionBuilder{}

	api := r.Group("/api/v1")

	UserRep := UserRepository.NewPgRepository(db)
	userCase := UserUseCase.NewUserUseCase(log.Info, log.Error, UserRep)
	UserHandler.NewRest(api.Group("/users"), userCase, &sessionBuilder, common.AuthRequired())

	TimetableRep := TimetableRepository.NewPgRepository(db)
	TimetableCase := TimetableUseCase.NewTimetableUseCase(TimetableRep)
	TimetableHandler.NewRest(api.Group("/timetable"), TimetableCase, &sessionBuilder, common.AuthRequired())

	app := App{
		config:   config,
		log:      log,
		route:    r,
		doneChan: make(chan bool, 1),
		db:       db,
	}

	return &app
}

func (a *App) Run() {

	srv := &http.Server{
		Addr:    a.config.Listen,
		Handler: a.route,
	}

	go func() {
		a.log.Info.Infof("Start listening on %s", a.config.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-a.doneChan:
	}
	a.log.Info.Info("Shutdown Server (timeout of 1 seconds) ...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		mes := fmt.Sprint("Server Shutdown:", err)
		a.log.Error.Fatal(mes)
	}

	<-ctx.Done()
	a.log.Info.Info("Server exiting")
}

func (a *App) Close() {
	a.doneChan <- true
}

func healthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Ok")
	}
}
