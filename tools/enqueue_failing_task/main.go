package main

import (
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	// Carrega as variáveis do ficheiro .env para obter REDIS_ADDR
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redisAddr := "localhost:6379"
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR not set in .env file")
	}

	// Cria um cliente Asynq
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// Cria uma tarefa de um tipo que não tem handler registado no worker
	// Isto garantirá que a tarefa falhe.
	task := asynq.NewTask("test:failure", []byte(`{"reason":"manual_test"}`), 
		asynq.MaxRetry(2),             // Tenta 2 vezes antes de arquivar
		asynq.Timeout(10*time.Second), // Define um timeout curto
	)

	// Enfileira a tarefa na fila "default"
	info, err := client.Enqueue(task, asynq.Queue("default"))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}

	log.Printf("Successfully enqueued failing task with ID: %s", info.ID)
}