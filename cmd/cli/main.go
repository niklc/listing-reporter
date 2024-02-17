package main

import (
	_ "github.com/joho/godotenv/autoload"

	listingreporter "github.com/niklc/listing-reporter/pkg/listing_reporter"
)

func main() {
	listingreporter.Run()
}
