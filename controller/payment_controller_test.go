package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPaymentController_CreatePayment(t *testing.T) {
	t.Run("should return 400 for invalid planId param", func(t *testing.T) {
		// PaymentController with nil deps — only tests param parsing, which happens before any dep is used
		ctrl := &PaymentController{}

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/payment/:planId", ctrl.CreatePayment)

		req := httptest.NewRequest("POST", "/payment/abc", bytes.NewReader([]byte(`{"name":"Test"}`)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON body", func(t *testing.T) {
		ctrl := &PaymentController{}

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/payment/:planId", ctrl.CreatePayment)

		req := httptest.NewRequest("POST", "/payment/1", bytes.NewReader([]byte("not valid json")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for missing required fields", func(t *testing.T) {
		ctrl := &PaymentController{}

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/payment/:planId", ctrl.CreatePayment)

		// Send valid JSON but missing required fields (e.g. no email, no password)
		req := httptest.NewRequest("POST", "/payment/1", bytes.NewReader([]byte(`{"name":"Test"}`)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestParsePaymentError(t *testing.T) {
	t.Run("should return CPF error for taxid message", func(t *testing.T) {
		msg, code := parsePaymentError("invalid taxId provided")
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Contains(t, msg, "CPF/CNPJ")
	})

	t.Run("should return email error for email message", func(t *testing.T) {
		msg, code := parsePaymentError("invalid email format")
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Contains(t, msg, "E-mail")
	})

	t.Run("should return phone error for cellphone message", func(t *testing.T) {
		msg, code := parsePaymentError("invalid cellphone number")
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Contains(t, msg, "Telefone")
	})

	t.Run("should return timeout error for deadline message", func(t *testing.T) {
		msg, code := parsePaymentError("context deadline exceeded")
		assert.Equal(t, http.StatusBadGateway, code)
		assert.Contains(t, msg, "indisponível")
	})

	t.Run("should return generic error for unknown message", func(t *testing.T) {
		msg, code := parsePaymentError("something unexpected happened")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Contains(t, msg, "Erro ao processar pagamento")
	})
}

func TestCleanString(t *testing.T) {
	t.Run("should remove non-numeric characters", func(t *testing.T) {
		assert.Equal(t, "12345678901", cleanString("123.456.789-01"))
	})

	t.Run("should handle empty string", func(t *testing.T) {
		assert.Equal(t, "", cleanString(""))
	})

	t.Run("should return only digits", func(t *testing.T) {
		assert.Equal(t, "11999887766", cleanString("(11) 99988-7766"))
	})
}
