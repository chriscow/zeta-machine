package main

import (
	z "zetamachine/pkg/lambda"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(z.RenderTile)
}
