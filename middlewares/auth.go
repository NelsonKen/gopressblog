package middlewares

import (
	"blog/models"
	"blog/services"
	"net/http"
	"strconv"
	"time"

	"github.com/fpay/gopress"
	"github.com/jinzhu/gorm"
)

// NewAuthMiddleware returns auth middleware.
// check cookies
func NewAuthMiddleware() gopress.MiddlewareFunc {
	return func(next gopress.HandlerFunc) gopress.HandlerFunc {
		return func(c gopress.Context) error {
			cookie, err := c.Cookie("uid")
			if err != nil {
				return c.Redirect(http.StatusFound, "/login?cookie=err")
			}

			if cookie.Value == "" {
				dropCookie(c, cookie)
				return c.Redirect(http.StatusFound, "/login?cookie=nil")
			}

			container := gopress.AppFromContext(c).Services

			dbs := container.Get(services.DBServerName).(*services.DBService)
			uid, err := strconv.Atoi(cookie.Value)
			if err != nil {
				dropCookie(c, cookie)
				return c.Redirect(http.StatusFound, "/login?cookie=invalid")
			}
			user := &models.User{}
			if dbs.ORM.First(user, uid).RecordNotFound() {
				dropCookie(c, cookie)
				return c.Redirect(http.StatusFound, "/login?cookie=nosuchuser")
			}

			messageNum := getMessageNum(dbs.ORM, user.ID)
			c.Set("messageNum", messageNum)
			c.Set("haveMessage", messageNum > 0)
			c.Set("user", user)

			return next(c)
		}
	}
}

func dropCookie(c gopress.Context, cookie *http.Cookie) {
	cookie.Expires = time.Now().Add(-1 * time.Second)
	c.SetCookie(cookie)
}

// getMessageNum
func getMessageNum(orm *gorm.DB, userID uint) int {
	var messageNum int
	orm.Model(&models.Message{}).Where("to_user_id = ? and readed = 0", userID).Count(&messageNum)
	return messageNum
}
