package redisutils

import (
	"BankTransactions/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func LPushTransfer(ctx context.Context, rClient *redis.Client, t *models.Transaction, telegramId int64) error {
	const op = "lib.redis.LPushTransfer"

	notification := map[string]interface{}{
		"to_user_id":   t.ToUserId,
		"from_user_id": t.FromUserId,
		"amount":       t.Amount,
		"message":      t.Message,
		"tg_id":        telegramId,
	}

	payload, _ := json.Marshal(notification)

	err := rClient.LPush(ctx, os.Getenv("REDIS_QUEUE_NAME"), payload).Err()
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}
