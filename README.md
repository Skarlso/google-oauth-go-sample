# Google OAuth Go Sample Project - Web application

This is a web application that demonstrates how to do Google Oauth to log-in an authenticate users.

# Installation

Simply `go get github.com/Skarlso/google-oauth-go-sample`.

# Setup

## Google

In order for the Google Authentication to work, you'll need developer credentials which the this application gathers from a file in the root directory called `creds.json`. The structure of this file should be like this:

```json
{
  "installed": {
    "client_id": "hash.apps.googleusercontent.com",
    "project_id": "random",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "secret",
    "redirect_uris": [
      "http://localhost"
    ]
  }
}
```

To obtain these credentials, please navigate to this site and follow the procedure to setup a new project: [Google Developer Console](https://console.developers.google.com/iam-admin/projects).

Once you have a new project, you need to create the above credentials. Navigate to the Project Page Credentials section
and create an Oauth Client ID. Select Desktop app and you should have your Client ID like the above JSON document.

## Dependencies

To gather all the libraries this project uses, simply execute from the root: `go get -v ./...`

# Running

To run it, simply build & run and navigate to http://127.0.0.1:9090/login, nothing else should be required.

```
go build
./oauth-sample
```
