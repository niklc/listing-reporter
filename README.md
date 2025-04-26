# Listing Reporter

The project scrapes [ss.lv](https://ss.lv) for listings and sends email alerts to subscribers about new matches, simplifying their search process.

The project is built upon AWS S3, DynamoDB, and Lambda for data storage, processing, and serverless computing, respectively. Additionally, it utilizes Google OAuth for user authentication and Gmail for email delivery.

## Email dependency

Email output requires Gmail API credentials as `credentials.json` and token as `token.json`. Guide on generating credentials [here](https://developers.google.com/gmail/api/quickstart/go). Token can be generated using `go run cmd/cli/main.go generate-token`.

##  Deployment

```bash
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap cmd/lambda/main.go
terraform init \
  -reconfigure \
  -backend-config="bucket=$BUCKET" \
  -backend-config="key=$KEY" \
  -backend-config="region=$REGION"
terraform apply
```
