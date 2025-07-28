package model

type AppSecrets struct {
	
	DBHost     string `json:"host_db"`
	DBPort     string `json:"port_db"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	
	RedisHost string `json:"redis_host"`
	RedisAddr string `json:"redis_addr"`
	RedisPort string `json:"redis_port"`
	RedisConf string `json:"redis_conf"`

	
	GeminiKey string `json:"gemini_key"`
	AIModel   string `json:"ai_model"`

	AdminNotificationEmail string `json:"admin_notification_email"`
	MonitorPollingInterval string `json:"monitor_polling_interval"`  // ex: "5m"
	NotifiedTaskSetKey     string `json:"notified_task_set_key"`
	NotifiedTaskTTL        string `json:"notified_task_ttl"`        // ex: "168h"
	QueuesToMonitor        string `json:"queues_to_monitor"`        // ex: "critical,default,low"
	SenderEmail            string `json:"sender_email"`
}