package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
)

type AuthorizerMiddleware struct {
	natsClient *nats.Conn
}

type AuthRequest struct {
	BotID     string `json:"bot_id"`
	ApiKey    string `json:"api_key"`
	Bytes     int64  `json:"bytes"`
	RequestID string `json:"request_id"`
}

type AuthResponse struct {
	Authorized    bool  `json:"authorized"`
	FuelRequired  int32 `json:"fuel_required"`
	FuelAvailable int32 `json:"fuel_available"`
}

func NewAuthorizerMiddleware(natsClient *nats.Conn) *AuthorizerMiddleware {
	return &AuthorizerMiddleware{
		natsClient: natsClient,
	}
}

// /////////////////////////////////////////
// Authorizer middleware
// This middleware is used to authorize a bot request by dispatching a request via NATS to the core service
// The core service will then authorize the request and return a response via NATS
// Authorizer returns the authorization middleware
func (f *MiddlewareFactory) Authorizer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestBody := AuthRequest{
				BotID:     f.config.BotID,
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

			response, err := f.natsClient.Request("v1.consumer.consume", requestBodyBytes, f.config.Timeout)
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
}
