package main

import (
	"fmt"
	"testing"
	"zetamachine/pkg/lambda"
	"zetamachine/pkg/zeta"
)

/*
Cross-origin resource sharing (CORS)
CORS is required to call your API from a webpage that isn’t hosted on the same domain.
To enable CORS for a REST API, set the Access-Control-Allow-Origin header in the
response object that you return from your function code.
*/
func TestHandler(t *testing.T) {
	tile := zeta.Tile{Zoom: 6, X: 0, Y: 0}
	b64, err := lambda.RenderTile(tile)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fmt.Println(b64)
	t.Fail()
}
