package middlewares

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	// Check for internal S2S token first
	internalToken := request.Headers["X-Internal-Token"]
	if internalToken == "" {
		internalToken = request.Headers["x-internal-token"]
	}
	expectedToken := os.Getenv("INTERNAL_AUTH_TOKEN")
	if internalToken != "" && expectedToken != "" && internalToken == expectedToken {
		return nil // Authorized as system
	}

	header := request.Headers["Authorization"]
	if header == "" {
		header = request.Headers["authorization"]
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return unauthorizedResponse("missing token")
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return unauthorizedResponse("invalid token")
	}

	return nil
}

func unauthorizedResponse(message string) *events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{"success": false, "error": message})
	return &events.APIGatewayProxyResponse{
		StatusCode: 401,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}
