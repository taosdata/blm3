package plugin

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/huskar-t/blm_demo/tools"
	"net/http"
	"strings"
)

const (
	UserKey     = "user"
	PasswordKey = "password"
)

func Auth(errHandler func(c *gin.Context, code int, err error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if len(auth) == 0 {
			errHandler(c, http.StatusUnauthorized, errors.New("auth needed"))
			c.Abort()
			return
		}
		auth = strings.TrimSpace(auth)
		if strings.HasPrefix(auth, "Basic") {
			b, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				errHandler(c, http.StatusUnauthorized, err)
				c.Abort()
				return
			}
			var user, password string
			sl := strings.Split(string(b), ":")

			if len(sl) == 2 {
				user = sl[0]
				password = sl[1]
			} else if len(sl) == 3 {
				if sl[2] == "aes" {
					encodeData, err := base64.StdEncoding.DecodeString(sl[0])
					if err != nil {
						errHandler(c, http.StatusUnauthorized, err)
						c.Abort()
						return
					}
					key, err := base64.StdEncoding.DecodeString(sl[1])
					if err != nil {
						errHandler(c, http.StatusUnauthorized, err)
						c.Abort()
						return
					}
					if len(key) != 16 {
						errHandler(c, http.StatusUnauthorized, errors.New("parse error"))
						c.Abort()
						return
					}
					authBytes, err := tools.AesDecrypt(encodeData, key)
					if err != nil {
						errHandler(c, http.StatusUnauthorized, err)
						c.Abort()
						return
					}
					a := strings.Split(string(authBytes), ":")
					if len(a) != 2 {
						errHandler(c, http.StatusUnauthorized, errors.New("parse error"))
						c.Abort()
						return
					}
					user = a[0]
					password = a[1]
				} else {
					errHandler(c, http.StatusUnauthorized, errors.New("unknown auth type"))
					c.Abort()
					return
				}
			} else {
				errHandler(c, http.StatusUnauthorized, errors.New("parse error"))
				c.Abort()
				return
			}
			c.Set(UserKey, user)
			c.Set(PasswordKey, password)
		}
	}
}

func RegisterGenerateAuth(r gin.IRouter) {
	r.GET("genauth/:user/:password/:key", func(c *gin.Context) {
		user := c.Param("user")
		password := c.Param("password")
		key := c.Param("key")
		if len(user) < 0 || len(user) > 24 || len(password) < 0 || len(password) > 24 || len(key) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		b := bytes.NewBufferString(user)
		b.WriteByte(':')
		b.WriteString(password)
		keyBytes := make([]byte, 16)
		maxLen := len(key)
		if maxLen > 16 {
			maxLen = 16
		}
		for i := 0; i < maxLen; i++ {
			keyBytes[i] = key[i]
		}
		d, err := tools.AesEncrypt(b.Bytes(), keyBytes)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		l1 := make([]byte, base64.StdEncoding.EncodedLen(len(d)))
		base64.StdEncoding.Encode(l1, d)
		l2 := make([]byte, base64.StdEncoding.EncodedLen(len(keyBytes)))
		base64.StdEncoding.Encode(l2, keyBytes)
		buf := bytes.Buffer{}
		buf.Write(l1)
		buf.WriteByte(':')
		buf.Write(l2)
		buf.WriteByte(':')
		buf.WriteString("aes")
		c.String(http.StatusOK, buf.String())
	})
}

func GetAuth(c *gin.Context) (user, password string, err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = errors.New("get auth error")
		}
	}()
	user = c.MustGet(UserKey).(string)
	password = c.MustGet(PasswordKey).(string)
	return
}
