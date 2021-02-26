package common

import (
	"bytes"
	"fmt"
	"github.com/apsdehal/go-logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

const ellipsisLength = 5 * 1024

func ellipsis(s string, l int) string {
	if len(s) > l {
		return s[:l] + "..."
	}
	return s
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				var err2 error
				switch v := err.(type) {
				case nil:
					err2 = fmt.Errorf("nil")
				case error:
					err2 = v
				case string:
					err2 = fmt.Errorf("%s", v)
				default:
					err2 = fmt.Errorf("%v", v)
				}
				err2 = fmt.Errorf("error: %s : %w ", string(debug.Stack()), err2)
				_ = c.AbortWithError(http.StatusInternalServerError, err2)
			}
		}()
		c.Next()
	}
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Content-type", "application/json")
		c.Next()
		if len(c.Errors) > 0 {
			var ret []interface{}
			for _, err := range c.Errors {
				switch err.Err.(type) {
				case Err:
					ret = append(ret, err.Err)
				default:
					ret = append(ret, NewErr(c.Writer.Status(), http.StatusText(c.Writer.Status())))
				}
			}
			c.JSON(0, ret)
		}
	}
}
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)

		// for response body log
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		message := fmt.Sprintf("\tExecution time: %s\tURL: %s\tMethod: %s\tHeaders: %s\tResponse status: %d\tX-Request-Id: %s",
			time.Since(startTime),
			c.Request.RequestURI,
			c.Request.Method,
			c.Request.Header,
			//ellipsis(string(body), ellipsisLength),
			c.Writer.Status(),
			c.Request.Header.Get("X-Request-Id"))
		//ellipsis(blw.body.String(), ellipsisLength))

		if c.Writer.Status() >= 400 && c.Writer.Status() < 500 {
			log.Warning(message)
		} else if c.Writer.Status() < 400 {
			log.Infof(message)
		} else {
			log.Error(message)
		}

		log.Debugf("\tRequest body: %s\tResponse body: %s", ellipsis(string(body), ellipsisLength), ellipsis(blw.body.String(), ellipsisLength))
	}
}

func ErrorLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// for response body log
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()
		for _, err := range c.Errors {
			e := fmt.Sprintf("ERROR:\t%s", err.Error())
			log.Error(e)
		}
	}
}

func Cors() gin.HandlerFunc {
	// Only for requests WITHOUT credentials, the literal value "*" can be specified
	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://127.0.0.1") ||
				strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "https://localhost") ||
				strings.HasPrefix(origin, "http://10-tka.pp.ua") ||
				strings.HasPrefix(origin, "https://10-tka.pp.ua") ||
				strings.HasPrefix(origin, "www.10-tka.pp.ua")
		},
		MaxAge: time.Hour,
	})
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("user_id") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "not authed",
			})
			c.Abort()
		}
	}
}
