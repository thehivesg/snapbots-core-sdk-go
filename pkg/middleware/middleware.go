package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"thehivesg/snapbots-core-sdk-go/pkg/config"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
)

type MiddlewareFactory struct {
	config     *config.Config
	natsClient *nats.Conn
}

func NewMiddlewareFactory(config *config.Config) *MiddlewareFactory {
	natsClient, err := nats.Connect(config.NATSURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	return &MiddlewareFactory{
		config:     config,
		natsClient: natsClient,
	}
}

func (a *MiddlewareFactory) Authorizer(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestBody := AuthRequest{
			BotID:     a.config.BotID,
			ApiKey:    c.Request().Header.Get("x-api-key"),
			RequestID: uuid.New().String(),
			Bytes:     c.Request().ContentLength,
		}

		// Set request id in the context
		c.Set("request_id", requestBody.RequestID)

		c.Logger().Info("Authorizing request")

		requestBodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			c.Logger().Error("Failed to marshal request body", err)
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		response, err := a.natsClient.Request("v1.consumer.consume", requestBodyBytes, 3*time.Second)
		if err != nil {
			c.Logger().Error("Failed to request authorization", err)
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		var authResponse AuthResponse
		err = json.Unmarshal(response.Data, &authResponse)
		if err != nil {
			c.Logger().Error("Failed to unmarshal authorization response", err)
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		c.Logger().Info("Auth Response: %+v\n", string(response.Data))

		if !authResponse.Authorized {
			c.Logger().Error("Unauthorized")
			return c.JSON(http.StatusUnauthorized, "Unauthorized")
		}

		c.Logger().Info("Authorized")
		c.Response().Header().Set("x-request-id", requestBody.RequestID)
		c.Response().Header().Set("x-bot-id", requestBody.BotID)
		c.Response().Header().Set("x-fuel-consumed", strconv.Itoa(int(authResponse.FuelRequired)))
		c.Response().Header().Set("x-fuel-available", strconv.Itoa(int(authResponse.FuelAvailable)))

		return next(c)
	}
}
