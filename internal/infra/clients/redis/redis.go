package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
)

type Client struct {
	client *redis.Client
}

type PaymentEntry struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	RequestedAt string  `json:"requested_at"`
}

func New() *Client {
	ctx := context.Background()

	// Configuração específica para Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379" // default Redis port
	}

	options := &redis.Options{
		Addr: redisURL,
		DB:   0, // database padrão

		PoolSize:     10,
		MinIdleConns: 3,
		MaxRetries:   3,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,

		MaxConnAge:  10 * time.Minute,
		IdleTimeout: 5 * time.Minute,
	}

	client := redis.NewClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(map[string]interface{}{
			"message": "error connecting to Redis",
			"error":   err,
			"url":     redisURL,
		})
	}

	log.Printf("Connected to Redis at %s", redisURL)

	return &Client{
		client: client,
	}
}

func (c *Client) Save(payload *entities.PaymentPayloadStorage) error {
	ctx := context.Background()

	entry := PaymentEntry{
		ID:          payload.ID,
		Amount:      payload.Amount,
		RequestedAt: payload.RequestedAt,
	}

	entryJSON, err := helpers.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error marshaling entry: %w", err)
	}

	key := fmt.Sprintf("%s:%s", string(payload.ProcessorProvider), payload.ID)

	if err := c.client.Set(ctx, key, entryJSON, 20*time.Minute).Err(); err != nil {
		return fmt.Errorf("error saving entry: %w", err)
	}

	return nil
}

func (c *Client) Retrieve(filters *entities.PaymentSummaryFilters) (*entities.PaymentResultStorage, error) {
	ctx := context.Background()

	result := &entities.PaymentResultStorage{
		entities.PaymentSummaryResponse{
			Default: entities.Summary{
				TotalRequests: 0,
				TotalAmount:   0,
			},
			Fallback: entities.Summary{
				TotalRequests: 0,
				TotalAmount:   0,
			},
		},
	}

	var (
		waitGroup sync.WaitGroup
		err       error
	)

	jobs := 2
	waitGroup.Add(jobs)

	go func() {
		defer waitGroup.Done()
		if processErr := c.setValues(ctx, entities.Default, result, filters); processErr != nil {
			err = processErr
		}
	}()

	go func() {
		defer waitGroup.Done()
		if processErr := c.setValues(ctx, entities.Fallback, result, filters); processErr != nil {
			err = processErr
		}
	}()

	waitGroup.Wait()

	return result, err
}

//nolint:cyclop,funlen // long but necessary
func (c *Client) setValues(
	ctx context.Context,
	processorProvider entities.ProcessorProvider,
	result *entities.PaymentResultStorage,
	filters *entities.PaymentSummaryFilters,
) error {
	pattern := string(processorProvider) + ":*"

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("error getting keys with pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()

	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("error executing pipeline: %w", err)
	}

	filterEnabled := filters != nil && filters.From != nil && filters.To != nil
	var utcFrom, utcTo time.Time

	if filterEnabled {
		utcFrom = filters.From.UTC()
		utcTo = filters.To.UTC()
	}

	var (
		totalRequests int
		totalAmount   float64
	)

	for _, cmd := range cmds {
		entryJSON, err := cmd.Result()
		if err != nil {
			continue
		}

		var entry PaymentEntry
		if err := helpers.Unmarshal([]byte(entryJSON), &entry); err != nil {
			continue
		}

		requestedAt, err := time.Parse(constants.DefaultTimeFormat, entry.RequestedAt)
		if err != nil {
			continue
		}

		if filterEnabled {
			if requestedAt.Before(utcFrom) || requestedAt.After(utcTo) {
				continue
			}
		}

		totalRequests++
		totalAmount += entry.Amount
	}

	roundNumber := 100
	roundFloat := float64(roundNumber)
	totalAmount = math.Round(totalAmount*roundFloat) / roundFloat

	if processorProvider == entities.Default {
		result.Default.TotalRequests += totalRequests
		result.Default.TotalAmount += totalAmount
	} else {
		result.Fallback.TotalRequests += totalRequests
		result.Fallback.TotalAmount += totalAmount
	}

	return nil
}
