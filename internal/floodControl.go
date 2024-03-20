package floodcontrol

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type FloodControlConfig struct {
	N uint //Время в секнудах
	K int  //Лимит сообщений
}

type RedisFloodControl struct {
	client *redis.Client
	config FloodControlConfig
}

func NewRedisFloodControl(client *redis.Client, config FloodControlConfig) *RedisFloodControl {
	return &RedisFloodControl{
		client: client,
		config: config,
	}
}

func (fc *RedisFloodControl) Check(ctx context.Context, userID int64) (bool, error) {

	var ttlSeconds = fc.config.N
	key := fmt.Sprintf("user:%v", userID)

	//Проверяем наличие ключа в БД
	exists, err := fc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	//Если ключа нет, то устанавливаем время жизни
	if exists == 0 {
		_, err = fc.client.Set(ctx, key, 0, time.Duration(ttlSeconds)*time.Second).Result()
		if err != nil {
			return false, err
		}
	}

	//Здесь мы увеличиваем счётчик запросов от пользователя на 1
	_, err = fc.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	//Получаем текущее значение по ключу
	count, err := fc.client.Get(ctx, key).Int()
	if err != nil {
		return false, err
	}

	//Проверяем превысили ли мы лимит
	if count > fc.config.K {
		return false, nil
	}
	return true, nil
}
