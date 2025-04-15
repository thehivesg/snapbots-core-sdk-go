package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
)

type TransactionMiddleware struct {
	natsClient *nats.Conn
}
type TransactionRequest struct {
	RequestID string `json:"request_id"`
}

func NewTransactionMiddleware(natsClient *nats.Conn) *TransactionMiddleware {
	return &TransactionMiddleware{
		natsClient: natsClient,
	}
}

// /////////////////////////////////////////
// Transaction middleware
// This middleware is used to confirm a transaction by dispatching a request via NATS to the core service
// The core service will then confirm the transaction and return a response via NATS
// Transaction returns the transaction middleware
func (f *MiddlewareFactory) Transaction() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get request id from the context
			requestID := c.Get("request_id")
			if requestID == nil {
				c.Logger().Error("Request ID not found in context")
				return c.JSON(http.StatusInternalServerError, "Request ID not found in context")
			}

			c.Logger().Info("Transaction middleware", requestID)

			transactionRequest := TransactionRequest{
				RequestID: requestID.(string),
			}

			transactionRequestBytes, err := json.Marshal(transactionRequest)
			if err != nil {
				c.Logger().Error("Failed to marshal transaction request", err)
				return c.JSON(http.StatusInternalServerError, err.Error())
			}

			err = f.natsClient.Publish("v1.transaction.confirmed", transactionRequestBytes)
			if err != nil {
				fmt.Printf("Failed to publish transaction request: %v\n", err)
				c.Logger().Error("Failed to publish transaction request", err)
				return c.JSON(http.StatusInternalServerError, err.Error())
			}

			return next(c)
		}
	}
}
