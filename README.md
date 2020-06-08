# MovieApp

MovieApp is a application that store lists of movies.



### Installation

MovieApp requires Go compiler to compile.

```sh
$ go build
```
### To run in your machine

- Change the firebase SDK snippet at end of templates/login.html to your project snippet.
- Get a TMDB API token with write privileges see [TMDB API](https://developers.themoviedb.org/4/auth/user-authorization-1).
- Get a [SendGrid](https://sendgrid.com/) API key.
- Get a service account key file from firebase see [Doc](https://firebase.google.com/docs/admin/setup?authuser=0).

Then set the env variables:
```sh
$ SENDGRID_API_KEY=<SendGrid API key>
$ TMDBTOKEN=<TMDB API token>
$ AUTHERCREDSPATH=<Path to service account key file>
$ MAILERADDR=<Email of your sender service>
```
