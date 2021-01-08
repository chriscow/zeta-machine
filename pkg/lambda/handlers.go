package lambda

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"zetamachine/pkg/zeta"

	"github.com/aws/aws-lambda-go/events"
)

func Compute(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var tile zeta.Tile

	if err := json.Unmarshal([]byte(req.Body), &tile); err != nil {
		log.Println("Error unmarshalling tile: ", err)
		return nil, err
	}

	algo := &zeta.Algo{}
	data := algo.Compute(tile.Min(), tile.Max(), nil)

	log.Println("Lambda received:", tile)

	buffer := bytes.NewBuffer(data)
	b64 := base64.StdEncoding.EncodeToString(buffer.Bytes())

	resp := &events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "text/text"},
		StatusCode: 200,
		Body:       b64,
	}

	return resp, nil
}
