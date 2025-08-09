package hazelcast

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/predicate"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
)

type Client struct {
	client    *hazelcast.Client
	clientMap *hazelcast.Map
}

func New(mapName string) *Client {
	gob.Register(PaymentEntry{})
	gob.Register([]PaymentEntry{})

	context := context.Background()

	config := hazelcast.NewConfig()

	config.Cluster.ConnectionStrategy.ReconnectMode = cluster.ReconnectModeOn

	if os.Getenv("HAZELCAST_URL") != "" {
		config.Cluster.Network.SetAddresses(os.Getenv("HAZELCAST_URL"))
	}

	client, err := hazelcast.StartNewClientWithConfig(context, config)
	if err != nil {
		log.Fatal(
			map[string]interface{}{
				"message": "error starting hazelcast",
				"error":   err,
			},
		)
	}

	myMap, err := client.GetMap(context, mapName)
	if err != nil {
		log.Fatal(
			map[string]interface{}{
				"message": "error getting map",
				"error":   err,
			},
		)
	}

	return &Client{
		client:    client,
		clientMap: myMap,
	}
}

func (c *Client) Save(payload *entities.PaymentPayloadStorage) error {
	context := context.Background()

	entry := PaymentEntry{
		ID:          payload.ID,
		Amount:      payload.Amount,
		RequestedAt: payload.RequestedAt,
	}

	key := fmt.Sprintf("%s:%s", string(payload.ProcessorProvider), payload.ID)

	if err := c.clientMap.Set(context, key, entry); err != nil {
		return fmt.Errorf("error saving entry: %w", err)
	}

	return nil
}

func (c *Client) Retrieve(filters *entities.PaymentSummaryFilters) (*entities.PaymentResultStorage, error) {
	context := context.Background()

	//nolint:govet
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
		err = c.setValues(context, entities.Default, result, filters)
	}()

	go func() {
		defer waitGroup.Done()
		err = c.setValues(context, entities.Fallback, result, filters)
	}()

	waitGroup.Wait()

	return result, err
}

//nolint:cyclop,funlen // long but necessary
func (c *Client) setValues(
	context context.Context,
	processorProvider entities.ProcessorProvider,
	result *entities.PaymentResultStorage,
	filters *entities.PaymentSummaryFilters,
) error {
	predicate := predicate.Like("__key", string(processorProvider)+":%%")

	entries, err := c.clientMap.GetEntrySetWithPredicate(context, predicate)
	if err != nil {
		return fmt.Errorf("error getting entries with predicate: %w", err)
	}

	for _, entry := range entries {
		//nolint:nestif // long but necessary
		if response, ok := entry.Value.(PaymentEntry); ok {
			requestedAt, err := time.Parse(constants.DefaultTimeFormat, response.RequestedAt)
			if err != nil {
				continue
			}

			filterEnabled := filters != nil && filters.From != nil && filters.To != nil

			inDateRange := filterEnabled && !requestedAt.Before(*filters.From) && !requestedAt.After(*filters.To)

			if processorProvider == entities.Default {
				if inDateRange {
					result.Default.TotalRequests++
					result.Default.TotalAmount += response.Amount

					continue
				}

				if filterEnabled {
					continue
				}

				result.Default.TotalRequests++
				result.Default.TotalAmount += response.Amount
			} else {
				if inDateRange {
					result.Fallback.TotalRequests++
					result.Fallback.TotalAmount += response.Amount

					continue
				}

				if filterEnabled {
					continue
				}

				result.Fallback.TotalRequests++
				result.Fallback.TotalAmount += response.Amount
			}
		}
	}

	roundNumber := 100
	roundFloat := float64(roundNumber)

	result.Default.TotalAmount = math.Round(result.Default.TotalAmount*roundFloat) / roundFloat
	result.Fallback.TotalAmount = math.Round(result.Fallback.TotalAmount*roundFloat) / roundFloat

	return nil
}
