# Listing Reporter

The project scrapes [ss.lv](https://ss.lv) for listings and sends email alerts to subscribers about new matches, simplifying their search process.

The project is built upon AWS S3, DynamoDB, and Lambda for data storage, processing, and serverless computing, respectively. Additionally, it utilizes Google OAuth for user authentication and Gmail for email delivery.

## Email dependency

Email output requires Gmail API credentials. Credentials can be generated <https://developers.google.com/gmail/api/quickstart/go>. Credentials must be saved as `credentials.json`.

Next `token.json` must be generated.

## Building Lambda

```bash
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap cmd/lambda/main.go
```
