package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/joho/godotenv/autoload"

	listingreporter "github.com/niklc/listing-reporter/pkg/listing_reporter"
	retrievalrules "github.com/niklc/listing-reporter/pkg/retrieval_rules"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("provide subcommand")
	}

	switch os.Args[1] {
	case "run":
		listingreporter.Run()
	case "get-rules":
		store := retrievalrules.NewRulesStore(session.Must(session.NewSession()))
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
		rule := retrievalrules.RetrievalRule{}
		err := json.Unmarshal([]byte(os.Args[2]), &rule)
		if err != nil {
			log.Fatal(err)
		}
		store := retrievalrules.NewRulesStore(session.Must(session.NewSession()))
		err = store.Put(rule)
		if err != nil {
			log.Fatal(err)
		}
	case "delete-rule":
		if len(os.Args) < 3 {
			log.Fatal("provide rule name")
		}
		store := retrievalrules.NewRulesStore(session.Must(session.NewSession()))
		err := store.Delete(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unknown command")
	}
}
