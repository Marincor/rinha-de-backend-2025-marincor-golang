package hazelcast

import (
	"context"
	"log"
	"os"
	"sync"

	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
)

type Client struct {
	client    *hazelcast.Client
	clientMap *hazelcast.Map
}

func New(mapName string) *Client {
	val := os.Environ()

	print(val)

	context := context.Background()

	config := hazelcast.NewConfig()

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

	if err := c.clientMap.Set(context, string(payload.ProcessorProvider), entry); err != nil {
		return fmt.Errorf("error saving entry: %w", err)
	}

	return nil
}

// Retrieve recupera e calcula o sumário de pagamentos baseado nos filtros
func (c *Client) Retrieve(payloadFilters *entities.PaymentSummaryFilters) (*entities.PaymentResultStorage, error) {
	context := context.Background()

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

	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		err = c.setValues(context, entities.Default, result)
	}()

	go func() {
		defer waitGroup.Done()
		err = c.setValues(context, entities.Fallback, result)
	}()

	waitGroup.Wait()

	return result, err
}
func (c *Client) setValues(
	context context.Context,
	processorProvider entities.ProcessorProvider,
	result *entities.PaymentResultStorage,
) error {
	values, err := c.clientMap.GetAll(context, string(processorProvider))
	if err != nil {
		return fmt.Errorf("error getting all entries: %w", err)
	}

	for _, value := range values {
		if processorProvider == entities.Default {
			result.Default.TotalRequests++
			floatValue, ok := value.Value.(float64)
			if ok {
				result.Default.TotalAmount += floatValue
			} else {
				log.Print(
					map[string]interface{}{
						"message": "error parsing value",
						"value":   value.Value,
					},
				)
			}

			if value.Key == "amount" {
				floatValue, ok := value.Value.(float64)
				if ok {
					result.Default.TotalAmount += floatValue
				} else {
					log.Print(
						map[string]interface{}{
							"message": "error parsing value",
							"value":   value.Value,
						},
					)
				}
			}
		} else {
			result.Fallback.TotalRequests++

			floatValue, ok := value.Value.(float64)
			if ok {
				result.Fallback.TotalAmount += floatValue
			} else {
				log.Print(
					map[string]interface{}{
						"message": "error parsing value",
						"value":   value.Value,
					},
				)
			}

			if value.Key == "amount" {
				floatValue, ok := value.Value.(float64)
				if ok {
					result.Fallback.TotalAmount += floatValue
				} else {
					log.Print(
						map[string]interface{}{
							"message": "error parsing value",
							"value":   value.Value,
						},
					)
				}
			}
		}
	}

	return nil
}
