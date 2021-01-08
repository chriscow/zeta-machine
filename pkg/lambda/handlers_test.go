package lambda

import (
	"log"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

/*
Cross-origin resource sharing (CORS)
CORS is required to call your API from a webpage that isn’t hosted on the same domain.
To enable CORS for a REST API, set the Access-Control-Allow-Origin header in the
response object that you return from your function code.
*/
func TestHandler(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Body: "{\"zoom\":0, \"x\":0, \"y\":1}",
	}

	resp, err := RenderTile(req)
	log.Println(err, resp)
	t.Fail()
}
