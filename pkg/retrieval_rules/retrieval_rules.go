package retrievalrules

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type RulesStore struct {
	dynamoSvc *dynamodb.DynamoDB
	tableName string
}

func NewRulesStore(awsSess *session.Session) *RulesStore {
	return &RulesStore{
		dynamoSvc: dynamodb.New(awsSess),
		tableName: "listing-reporter",
	}
}

type RetrievalRule struct {
	Name    string
	Email   string
	Url     string
	Filters Filters
	Cutoffs []string
}

type Filters struct {
	Price         *RangeFilter[float64]
	Rooms         *RangeFilter[int]
	Area          *RangeFilter[float64]
	Floor         *RangeFilter[int]
	IsNotTopFloor *bool
}

type RangeFilter[T int | float64] struct {
	From *T
	To   *T
}

func (r *RulesStore) Get() ([]RetrievalRule, error) {
	res, err := r.dynamoSvc.Scan(&dynamodb.ScanInput{TableName: &r.tableName})
	if err != nil {
		return nil, err
	}
	rules := make([]RetrievalRule, len(res.Items))
	for i, item := range res.Items {
		err = dynamodbattribute.UnmarshalMap(item, &rules[i])
		if err != nil {
			return nil, err
		}
	}
	return rules, nil
}

func (r *RulesStore) Put(rule RetrievalRule) error {
	av, err := dynamodbattribute.MarshalMap(rule)
	if err != nil {
		return err
	}
	_, err = r.dynamoSvc.PutItem(&dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      av,
	})
	return err
}

func (r *RulesStore) Delete(name string) error {
	_, err := r.dynamoSvc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key:       map[string]*dynamodb.AttributeValue{"Name": {S: &name}},
	})
	return err
}
