GitHub Pull Request Checker
----

[![GoDoc][1]][2] [![License: MIT][3]][4] [![Release][5]][6] [![Build Status][7]][8] [![Go Report Card][13]][14] [![Code Climate][19]][20] [![BCH compliance][21]][22]

[1]: https://godoc.org/github.com/evalphobia/github-pr-checker?status.svg
[2]: https://godoc.org/github.com/evalphobia/github-pr-checker
[3]: https://img.shields.io/badge/License-MIT-blue.svg
[4]: LICENSE.md
[5]: https://img.shields.io/github/release/evalphobia/github-pr-checker.svg
[6]: https://github.com/evalphobia/github-pr-checker/releases/latest
[7]: https://travis-ci.org/evalphobia/github-pr-checker.svg?branch=master
[8]: https://travis-ci.org/evalphobia/github-pr-checker
[9]: https://coveralls.io/repos/evalphobia/github-pr-checker/badge.svg?branch=master&service=github
[10]: https://coveralls.io/github/evalphobia/github-pr-checker?branch=master
[11]: https://codecov.io/github/evalphobia/github-pr-checker/coverage.svg?branch=master
[12]: https://codecov.io/github/evalphobia/github-pr-checker?branch=master
[13]: https://goreportcard.com/badge/github.com/evalphobia/github-pr-checker
[14]: https://goreportcard.com/report/github.com/evalphobia/github-pr-checker
[15]: https://img.shields.io/github/downloads/evalphobia/github-pr-checker/total.svg?maxAge=1800
[16]: https://github.com/evalphobia/github-pr-checker/releases
[17]: https://img.shields.io/github/stars/evalphobia/github-pr-checker.svg
[18]: https://github.com/evalphobia/github-pr-checker/stargazers
[19]: https://codeclimate.com/github/evalphobia/github-pr-checker/badges/gpa.svg
[20]: https://codeclimate.com/github/evalphobia/github-pr-checker
[21]: https://bettercodehub.com/edge/badge/evalphobia/github-pr-checker?branch=master
[22]: https://bettercodehub.com/


`github-pr-checker` is a tool to help pull request to post comment and set assignees automatically.


# What's for?

This tool checks pull request event and gets changed files list through GitHub webhook.
You can set regexp rules for the changed files and this tool post comments, assignees and reviwers from the rules.

# Quick Usage

At first, install golang.
And gets dependensies.

```bash
$ make init
```

Then create config.yml and modify it for your own needs.

```bash
$ cp config.example.yml config.yml
$ vim config.yml
```


## Serverless for AWS Lambda

Create serverless.yml

```bash
$ cp serverless.example.yml serverless.yml

# see https://serverless.com/framework/docs/providers/aws/guide/serverless.yml/
$ vim serverless.yml
```

Then deploy it!

```bash
# $ AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... make deploy
$ make deploy

...

Serverless: Stack update finished...
Service Information
service: serverless-github-pr-review
stage: dev
region: ap-northeast-1
stack: serverless-github-pr-review-dev
api keys:
  None
endpoints:
  POST - https://abcdefg.execute-api.ap-northeast-1.amazonaws.com/dev/func
functions:
  func: serverless-github-pr-review-dev-func
layers:
  None

Stack Outputs
FuncLambdaFunctionQualifiedArn: arn:aws:lambda:ap-northeast-1:...:function:serverless-github-pr-review-dev-func:1
ServiceEndpoint: https://abcdefg.execute-api.ap-northeast-1.amazonaws.com/dev
ServerlessDeploymentBucketName: serverless-github-pr-review-dev-serverlessdeploymentbuck-123456
```

On the above example, endpoint URL is `https://abcdefg.execute-api.ap-northeast-1.amazonaws.com/dev`.
Set the URL to GitHub repository's webhook Payload URL.


## HTTP server

```bash
# $ GITHUB_PR_API_TOKEN=... go run ./cmd/httpserver/main.go
$ GITHUB_PR_HTTP_PORT=3000 go run ./cmd/httpserver/main.go
```

If you debug in local machine, [smee.io](https://smee.io/) will be really helpful.

```bash
# console #1
$ GITHUB_PR_HTTP_PORT=3000 go run ./cmd/httpserver/main.go

# console #2
# $ npm install -g smee-client
$ smee -u https://smee.io/123456
```

# GitHub's Webhook setting

Go to `https://github.com/<owner>/<repo name>/settings/hooks/new` and set data.

|Name|Data|
|:--|:--|
| `Payload URL` | Set HTTP server's URL or Serverless Endpoint. |
| `Content type` | `application/json` |
| `Secret` | Keep blank. If you set it, then set `webhook_secret` in `config.yml` |
| `Which events would you like to trigger this webhook?` | Use `Let me select individual events.` and check `Pull requests` |
| `Active` | Check! |


# Environment variables

|Name|Description|
|:--|:--|
| `GITHUB_PR_HTTP_PORT` | HTTP server's listen port. (default: `3000`) |
| `GITHUB_PR_CONFIG_FILE` | Config setting yaml file. (default: `config.yml`) |
| `GITHUB_PR_API_TOKEN` | GitHub's personal access token. |
| `GITHUB_PR_API_TOKEN_KMS` | GitHub's personal access token encrypted by AWS KMS. |
