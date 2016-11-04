# Google OAuth Go Sample Project - Web application

Web based version of [GoProgressQuest](https://github.com/Skarlso/goprogressquest).

# Installation

Simply `go get github.com/Skarlso/google-oauth-go-sample`.

# Setup

## Google

In order for the Google Authentication to work, you'll need developer credentials which the this application gathers from a file in the root directory called `creds.json`. The structure of this file should be like this:

```json
{
  "cid":"hash.apps.googleusercontent.com",
  "csecret":"somesecrethash"
}
```

To obtain these credentials, please navigate to this site and follow the procedure to setup a new project: [Google Developer Console](https://console.developers.google.com/iam-admin/projects).

## Dependencies

To gather all the libraries this project uses, simply execute from the root: `go get -v ./...`

# Running

To run it, simply build & run and navigate to http://127.0.0.1:9090/login, nothing else should be required.

```
go build
./google-oauth-go-sample
```
