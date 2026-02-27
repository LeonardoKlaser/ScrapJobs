package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupUserController() (UserController, *mocks.MockUserRepository) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsercase(mockRepo)
	ctrl := NewUserController(uc)
	return ctrl, mockRepo
}

func TestUserController_SignIn(t *testing.T) {
	t.Run("should sign in successfully and set cookie", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		hashedPw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com", Password: string(hashedPw)}
		mockRepo.On("GetUserByEmail", "test@test.com").Return(user, nil).Once()

		t.Setenv("JWTTOKEN", "test-secret-key-that-is-at-least-32-chars")

		body, _ := json.Marshal(map[string]string{"email": "test@test.com", "password": "password123"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/login", ctrl.SignIn)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "Authorization" {
				found = true
				assert.NotEmpty(t, cookie.Value)
				break
			}
		}
		assert.True(t, found, "Authorization cookie should be set")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for wrong password", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		hashedPw, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com", Password: string(hashedPw)}
		mockRepo.On("GetUserByEmail", "test@test.com").Return(user, nil).Once()

		body, _ := json.Marshal(map[string]string{"email": "test@test.com", "password": "wrongpassword"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/login", ctrl.SignIn)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Invalid E-mail or Password", resp["error"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for user not found (timing-safe)", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		// Return zero-value user (Id == 0) to simulate user not found
		mockRepo.On("GetUserByEmail", "notfound@test.com").Return(model.User{}, nil).Once()

		body, _ := json.Marshal(map[string]string{"email": "notfound@test.com", "password": "anypassword"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/login", ctrl.SignIn)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Invalid E-mail or Password", resp["error"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for empty email", func(t *testing.T) {
		ctrl, _ := setupUserController()

		body, _ := json.Marshal(map[string]string{"email": "", "password": "password123"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/login", ctrl.SignIn)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserController_Logout(t *testing.T) {
	t.Run("should logout and clear cookie", func(t *testing.T) {
		ctrl, _ := setupUserController()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/logout", ctrl.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Logout successful", resp["message"])

		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "Authorization" {
				found = true
				assert.Equal(t, "", cookie.Value)
				assert.True(t, cookie.MaxAge < 0, "Cookie MaxAge should be negative to clear it")
				break
			}
		}
		assert.True(t, found, "Authorization cookie should be set (cleared)")
	})
}

func TestUserController_ValidateCheckout(t *testing.T) {
	t.Run("should return both false when neither exists", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		mockRepo.On("CheckUserExists", "new@test.com", "12345678901").Return(false, false, nil).Once()

		body, _ := json.Marshal(map[string]string{"email": "new@test.com", "tax": "123.456.789-01"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/validate", ctrl.ValidateCheckout)

		req := httptest.NewRequest("POST", "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]bool
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["email_exists"])
		assert.False(t, resp["tax_exists"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return email_exists true when email exists", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		mockRepo.On("CheckUserExists", "existing@test.com", "12345678901").Return(true, false, nil).Once()

		body, _ := json.Marshal(map[string]string{"email": "existing@test.com", "tax": "123.456.789-01"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/validate", ctrl.ValidateCheckout)

		req := httptest.NewRequest("POST", "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]bool
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["email_exists"])
		assert.False(t, resp["tax_exists"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for missing required fields", func(t *testing.T) {
		ctrl, _ := setupUserController()

		body, _ := json.Marshal(map[string]string{"email": "test@test.com"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/validate", ctrl.ValidateCheckout)

		req := httptest.NewRequest("POST", "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserController_UpdateProfile(t *testing.T) {
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should update profile successfully", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		mockRepo.On("UpdateUserProfile", 1, "New Name", (*string)(nil), (*string)(nil)).Return(nil).Once()

		body, _ := json.Marshal(map[string]string{"user_name": "New Name"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/profile", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.UpdateProfile(c)
		})

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Perfil atualizado com sucesso", resp["message"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for missing name", func(t *testing.T) {
		ctrl, _ := setupUserController()

		body, _ := json.Marshal(map[string]string{"user_name": ""})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/profile", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.UpdateProfile(c)
		})

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 401 when user not in context", func(t *testing.T) {
		ctrl, _ := setupUserController()

		body, _ := json.Marshal(map[string]string{"user_name": "Name"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/profile", ctrl.UpdateProfile)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUserController_ChangePassword(t *testing.T) {
	t.Run("should change password successfully", func(t *testing.T) {
		ctrl, mockRepo := setupUserController()

		hashedPw, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com", Password: string(hashedPw)}

		mockRepo.On("UpdateUserPassword", 1, mock.AnythingOfType("string")).Return(nil).Once()

		body, _ := json.Marshal(map[string]string{"old_password": "oldpassword", "new_password": "newpassword123"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/password", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.ChangePassword(c)
		})

		req := httptest.NewRequest("PUT", "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Senha alterada com sucesso", resp["message"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for wrong old password", func(t *testing.T) {
		ctrl, _ := setupUserController()

		hashedPw, _ := bcrypt.GenerateFromPassword([]byte("correctold"), bcrypt.DefaultCost)
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com", Password: string(hashedPw)}

		body, _ := json.Marshal(map[string]string{"old_password": "wrongold", "new_password": "newpassword123"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/password", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.ChangePassword(c)
		})

		req := httptest.NewRequest("PUT", "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for short new password", func(t *testing.T) {
		ctrl, _ := setupUserController()

		hashedPw, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com", Password: string(hashedPw)}

		body, _ := json.Marshal(map[string]string{"old_password": "oldpassword", "new_password": "short"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/password", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.ChangePassword(c)
		})

		req := httptest.NewRequest("PUT", "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 401 when user not in context", func(t *testing.T) {
		ctrl, _ := setupUserController()

		body, _ := json.Marshal(map[string]string{"old_password": "old", "new_password": "newpassword123"})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/password", ctrl.ChangePassword)

		req := httptest.NewRequest("PUT", "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
