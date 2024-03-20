package main

import (
	"context"
	"encoding/json"
	floodcontrol "flood-control-task/internal"

	"fmt"
	"io"
	"log"
	"os"

	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

type ConfigRedis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	db       int    `yaml:"db"`
}

func initConfig() (*ConfigRedis, error) {
	var config ConfigRedis

	file, err := os.Open("configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %v", err)
	}

	return &config, nil
}

type ConfigJSON struct {
	K int
	N uint
}

func loadConfigJSON(filename string) (ConfigJSON, error) {
	var config ConfigJSON

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func main() {

	configRedis, err := initConfig()
	if err != nil {
		log.Fatal("error initializing config:", err)
	}
	//Подключаемся к редис и проверяем работу соединения
	cleint := redis.NewClient(&redis.Options{
		Addr:     configRedis.Addr,
		Password: configRedis.Password,
		DB:       configRedis.db,
	})

	configJSON, err := loadConfigJSON("configs/config.json")
	if err != nil {
		fmt.Println("error loading config.json", err)
	}

	ctx := context.Background()

	_, err = cleint.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error connecting to Redis", err)
		return
	}

	time.Sleep(2 * time.Second)
	fmt.Printf("Время в секундах: %v\n", configJSON.N)
	fmt.Printf("Лимит сообщений: %v\n", configJSON.K)

	configFloodControl := floodcontrol.FloodControlConfig{
		N: configJSON.N, //Время в секундах!
		K: configJSON.K, //Наш лимит сообщений
	}

	fc := floodcontrol.NewRedisFloodControl(cleint, configFloodControl)

	//Проверяем RedisFloodControl для пользовтелья с ID 1
	userID := int64(2)
	for i := 0; i < 10; i++ {
		ok, err := fc.Check(ctx, userID)
		if err != nil {
			fmt.Println("Error checking flood control:", err)
			return
		}
		if ok {
			fmt.Println("ОК")
		} else {
			fmt.Println("Лимит сообщений превышен")
		}
		time.Sleep(time.Second)
	}

}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}
