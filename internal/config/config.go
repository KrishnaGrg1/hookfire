package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DB_URL string
	REDIS_URL string
	WORKER_COUNT int
}

func Load()*Config{

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	workerStr:=getEnv("WORKER_COUNT","10")
	workerCount,err:=strconv.Atoi(workerStr)
	if err!=nil{
		log.Fatal("Error parsing WORKER_COUNT")
	}
	return &Config{
		Port: getEnv("PORT","8080"),
		DB_URL: getEnv("GOOSE_DBSTRING",""),
		REDIS_URL: getEnv("REDIS_URL",""),
		WORKER_COUNT: workerCount,
	}
	
}

func getEnv(key ,fallback string)string{
	v:=os.Getenv(key)
	if v==""{
		return fallback
	}
	return v
}
