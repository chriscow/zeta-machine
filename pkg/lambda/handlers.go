package lambda

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"zetamachine/pkg/zeta"

	"github.com/aws/aws-lambda-go/events"
)

func Compute(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var tile zeta.Tile

	if err := json.Unmarshal([]byte(req.Body), &tile); err != nil {
		log.Println("Error unmarshalling tile: ", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	jsonb, err := zeta.ComputeRequest(ctx, []byte(req.Body), nil)
	if err != nil {
		log.Println("Error computing tile: ", err)
		return nil, err
	}

	resp := &events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: 200,
		Body:       string(jsonb),
	}

	return resp, nil
}
