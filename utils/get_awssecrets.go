package utils

import (
	"context"
	"encoding/json"
	"web-scrapper/model"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/joho/godotenv"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetAppSecrets(secretName string) (*model.AppSecrets, error){
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil{
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := svc.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, err
	}

	var secrets model.AppSecrets
	err = json.Unmarshal([]byte(*result.SecretString), &secrets)
	if err != nil {
		return nil, err
	}

	return &secrets, nil
}


// MonitorConfig detém a configuração processada e pronta para uso pelo serviço de monitoramento.
type MonitorConfig struct {
	RedisAddr              string
	PollingInterval        time.Duration
	AdminNotificationEmail string
	NotifiedTaskSetKey     string
	NotifiedTaskTTL        time.Duration
	QueuesToMonitor        []string
	SenderEmail            string
}

// LoadMonitorConfig carrega a configuração para o archive-monitor.
// Ele busca segredos da AWS se configurado, caso contrário, usa variáveis de ambiente locais.
func LoadMonitorConfig() (*MonitorConfig, error) {
	if os.Getenv("GIN_MODE") != "release" {
		godotenv.Load()
	}

	var secrets *model.AppSecrets
	secretName := os.Getenv("APP_SECRET_NAME")
	if secretName != "" {
		var err error
		secrets, err = GetAppSecrets(secretName)
		if err != nil {
			return nil, fmt.Errorf("failed to get secrets from AWS Secrets Manager: %w", err)
		}
	}

	// Função auxiliar para obter valor do secret ou fallback para env var
	getVal := func(envKey, defaultVal string) string {
		if secrets != nil {
			switch envKey {
			case "REDIS_ADDR":
				if secrets.RedisAddr != "" { return secrets.RedisAddr }
			case "MONITOR_POLLING_INTERVAL":
				if secrets.MonitorPollingInterval != "" { return secrets.MonitorPollingInterval }
			case "ADMIN_NOTIFICATION_EMAIL":
				if secrets.AdminNotificationEmail != "" { return secrets.AdminNotificationEmail }
			case "NOTIFIED_TASK_SET_KEY":
				if secrets.NotifiedTaskSetKey != "" { return secrets.NotifiedTaskSetKey }
			case "NOTIFIED_TASK_TTL":
				if secrets.NotifiedTaskTTL != "" { return secrets.NotifiedTaskTTL }
			case "QUEUES_TO_MONITOR":
				if secrets.QueuesToMonitor != "" { return secrets.QueuesToMonitor }
			case "SENDER_EMAIL":
				if secrets.SenderEmail != "" { return secrets.SenderEmail }
			}
		}

		val := os.Getenv(envKey)
		if val != "" {
			return val
		}

		return defaultVal
	}

	// Carrega e processa cada variável de configuração
	pollingIntervalStr := getVal( "MONITOR_POLLING_INTERVAL", "5m")
	pollingInterval, err := time.ParseDuration(pollingIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MONITOR_POLLING_INTERVAL format: %w", err)
	}

	notifiedTaskTTLStr := getVal( "NOTIFIED_TASK_TTL", "168h")
	notifiedTaskTTL, err := time.ParseDuration(notifiedTaskTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid NOTIFIED_TASK_TTL format: %w", err)
	}

	queuesStr := getVal( "QUEUES_TO_MONITOR", "default")

	cfg := &MonitorConfig{
		RedisAddr:              getVal("REDIS_ADDR", ""),
		PollingInterval:        pollingInterval,
		AdminNotificationEmail: getVal("ADMIN_NOTIFICATION_EMAIL", ""),
		NotifiedTaskSetKey:     getVal("NOTIFIED_TASK_SET_KEY", "scrapjobs:notified_archived_tasks"),
		NotifiedTaskTTL:        notifiedTaskTTL,
		QueuesToMonitor:        strings.Split(queuesStr, ","),
		SenderEmail:            getVal("SENDER_EMAIL", "noreply@scrapjobs.com.br"),
	}

	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("redis address (REDIS_ADDR) must be configured")
	}
	if cfg.AdminNotificationEmail == "" {
		return nil, fmt.Errorf("admin notification email (ADMIN_NOTIFICATION_EMAIL) must be configured")
	}

	return cfg, nil
}


