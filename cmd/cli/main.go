package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/joho/godotenv/autoload"
	reporter "github.com/niklc/listing-reporter/internal"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("provide subcommand")
	}

	switch os.Args[1] {
	case "run":
		start := time.Now()
		reporter.Run()
		fmt.Printf("Execution time: %s\n", time.Since(start))
	case "generate-token":
		file, err := os.Open("credentials.json")
		if err != nil {
			log.Fatalf("unable to open credentials file: %v", err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("unable to read credentials file: %v", err)
		}

		reporter.GetAndSaveToken(data)
	case "get-rules":
		store := reporter.NewRulesStore(session.Must(session.NewSession()))
		rules, err := store.Get()
		if err != nil {
			log.Fatal(err)
		}
		for _, rule := range rules {
			filters, err := json.Marshal(rule)
			if err != nil {
				log.Println("print rule failed: ", err)
			}

			fmt.Println(string(filters))
		}
	case "put-rule":
		if len(os.Args) < 3 {
			log.Fatal("provide rule")
		}
		rule := reporter.RetrievalRule{}
		err := json.Unmarshal([]byte(os.Args[2]), &rule)
		if err != nil {
			log.Fatal(err)
		}
		store := reporter.NewRulesStore(session.Must(session.NewSession()))
		err = store.Put(rule)
		if err != nil {
			log.Fatal(err)
		}
	case "delete-rule":
		if len(os.Args) < 3 {
			log.Fatal("provide rule name")
		}
		store := reporter.NewRulesStore(session.Must(session.NewSession()))
		err := store.Delete(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unknown command")
	}
}
