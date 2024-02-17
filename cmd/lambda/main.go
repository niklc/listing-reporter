package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	listingreporter "github.com/niklc/listing-reporter/pkg/listing_reporter"
)

func HandleRequest() {
	listingreporter.Run()
}

func main() {
	lambda.Start(HandleRequest)
}
