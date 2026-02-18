package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"web-scrapper/logging"

	"github.com/gin-gonic/gin"
)

// VerifyWebhookSecret valida o secret passado como query param na URL do webhook.
// A AbacatePay envia: POST /webhook?webhookSecret=<seu_secret>
func VerifyWebhookSecret(c *gin.Context) bool {
	webhookSecret := c.Query("webhookSecret")
	expectedSecret := os.Getenv("ABACATEPAY_WEBHOOK_SECRET")

	if expectedSecret == "" {
		logging.Logger.Error().Msg("ABACATEPAY_WEBHOOK_SECRET não está configurado no ambiente!")
		return false
	}

	if webhookSecret != expectedSecret {
		logging.Logger.Warn().Str("received_secret", webhookSecret).Msg("Secret do webhook inválido recebido na query string")
		return false
	}
	return true
}

// VerifyWebhookHMACSignature valida a assinatura HMAC-SHA256 do corpo do webhook.
// A AbacatePay usa sua PUBLIC KEY como chave HMAC e envia a assinatura em base64 no header X-Webhook-Signature.
// Comparação deve ser feita em base64 (não decodificando o header).
func VerifyWebhookHMACSignature(c *gin.Context, rawBody []byte) bool {
	signatureFromHeader := c.GetHeader("X-Webhook-Signature")
	// A chave pública da AbacatePay (fixa por ambiente, configurada via env var)
	publicKey := os.Getenv("ABACATEPAY_PUBLIC_KEY")

	if signatureFromHeader == "" {
		logging.Logger.Warn().Msg("Cabeçalho X-Webhook-Signature ausente na requisição do webhook")
		return false
	}
	if publicKey == "" {
		logging.Logger.Error().Msg("ABACATEPAY_PUBLIC_KEY não está configurado no ambiente!")
		return false
	}

	// Calcula HMAC-SHA256 do corpo usando a public key como chave
	mac := hmac.New(sha256.New, []byte(publicKey))
	mac.Write(rawBody)
	expectedSigBytes := mac.Sum(nil)

	// Converte para base64 para comparar com o header (que também é base64)
	expectedSigBase64 := base64.StdEncoding.EncodeToString(expectedSigBytes)

	// Comparação segura contra timing attacks
	expectedBuf := []byte(expectedSigBase64)
	receivedBuf := []byte(signatureFromHeader)

	if len(expectedBuf) != len(receivedBuf) {
		logging.Logger.Warn().
			Str("received_signature", signatureFromHeader).
			Str("expected_signature", expectedSigBase64).
			Msg("Assinatura HMAC do webhook inválida (tamanhos diferentes)")
		return false
	}

	if !hmac.Equal(expectedBuf, receivedBuf) {
		logging.Logger.Warn().
			Str("received_signature", signatureFromHeader).
			Str("expected_signature", expectedSigBase64).
			Msg("Assinatura HMAC do webhook inválida")
		return false
	}

	return true
}

// WebhookAuthMiddleware valida a autenticidade do webhook da AbacatePay em duas camadas:
// 1. Secret na query string (?webhookSecret=...)
// 2. Assinatura HMAC no header X-Webhook-Signature
func WebhookAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Camada 1: valida o secret na URL
		if !VerifyWebhookSecret(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook secret"})
			return
		}

		// Lê o corpo bruto para validação HMAC (deve ser feito antes de qualquer parsing)
		rawBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logging.Logger.Error().Err(err).Msg("Erro ao ler corpo da requisição do webhook para verificação HMAC")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error reading body"})
			return
		}
		// Restaura o body para que o handler possa lê-lo novamente
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		// Camada 2: valida a assinatura HMAC
		if !VerifyWebhookHMACSignature(c, rawBody) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
			return
		}

		logging.Logger.Debug().Msg("Autenticação do webhook AbacatePay bem-sucedida")
		c.Next()
	}
}
