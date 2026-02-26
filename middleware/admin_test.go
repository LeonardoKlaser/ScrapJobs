package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireAdmin_IsAdmin(t *testing.T) {
	router := gin.New()
	router.GET("/admin", func(ctx *gin.Context) {
		ctx.Set("user", model.User{Id: 1, Name: "Admin", Email: "admin@test.com", IsAdmin: true})
		ctx.Next()
	}, RequireAdmin(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_AdminEmail(t *testing.T) {
	os.Setenv("ADMIN_EMAIL", "admin@scrapjobs.com")
	defer os.Unsetenv("ADMIN_EMAIL")

	router := gin.New()
	router.GET("/admin", func(ctx *gin.Context) {
		ctx.Set("user", model.User{Id: 2, Name: "Admin Email", Email: "admin@scrapjobs.com", IsAdmin: false})
		ctx.Next()
	}, RequireAdmin(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_NotAdmin(t *testing.T) {
	os.Setenv("ADMIN_EMAIL", "admin@scrapjobs.com")
	defer os.Unsetenv("ADMIN_EMAIL")

	router := gin.New()
	router.GET("/admin", func(ctx *gin.Context) {
		ctx.Set("user", model.User{Id: 3, Name: "Regular User", Email: "user@test.com", IsAdmin: false})
		ctx.Next()
	}, RequireAdmin(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_NoUserInContext(t *testing.T) {
	router := gin.New()
	router.GET("/admin", RequireAdmin(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
