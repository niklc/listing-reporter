Email output requires Gmail API credentials. Credentials can be generated https://developers.google.com/gmail/api/quickstart/go . Credentials must be saved in `credentials.json`.

First run must be done locally so that OAuth 2.0 token is requested from Google and saved to `token.json`.

"to" email address in `listing_reporter.go` and filters in `filters.json` must be set before building image.