package lambda

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"log"
	"zetamachine/pkg/zeta"

	"github.com/aws/aws-lambda-go/events"
)

type RenderTileResponse struct {
	Tile  zeta.Tile `json:"tile"`
	Image string    `json:"image"`
}

func RenderTile(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var tile zeta.Tile
	if err := json.Unmarshal([]byte(req.Body), &tile); err != nil {
		log.Println("Error unmarshalling tile: ", err)
		return nil, err
	}

	result := RenderTileResponse{
		Tile: tile,
	}

	log.Println("Lambda received:", tile)
	img, err := zeta.NewTile(tile.Zoom, tile.X, tile.Y)
	if err != nil {
		log.Println("Error while calling NewTile: ", err)
		return nil, err
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, *img); err != nil {
		log.Println("Unable to encode image: ", err)
		return nil, err
	}

	result.Image = base64.StdEncoding.EncodeToString(buffer.Bytes())

	b, err := json.Marshal(result)
	if err != nil {
		log.Println("Error marshaling result: ", err)
		return nil, err
	}

	resp := &events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: 200,
		Body:       string(b),
	}

	return resp, nil
}
