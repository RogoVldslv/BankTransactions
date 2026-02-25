package redisutils

import (
	"BankTransactions/internal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func LPushTransfer(ctx context.Context, rClient *redis.Client, t *models.Transaction) error {
	const op = "lib.redis.LPushTransfer"

	notification := map[string]interface{}{
		"to_user_id":   t.ToUserId,
		"from_user_id": t.FromUserId,
		"amount":       t.Amount,
		"message":      t.Message,
	}

	payload, _ := json.Marshal(notification)

	err := rClient.LPush(ctx, "tg_notifications", payload).Err()
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}
