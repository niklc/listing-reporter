package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	reporter "github.com/niklc/listing-reporter/internal"
)

func HandleRequest() {
	reporter.Run()
}

func main() {
	lambda.Start(HandleRequest)
}
