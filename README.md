# blog-aggregator (gator)

Gator is a CLI-based RSS feed agreggator microservice built with go.

Tools used:

- postgresql
- sqlc
- goose
- psql

## Purpose

The gator service parse data from RSS feeds. The service can be configured to continuosly fetch data in a specific interval and display the new information.

The commands interact with a PostgreSQL database and allow users to register, log in, add feeds, follow/unfollow feeds, and browse content.

## Requirements

To set up the project, you'll Postgresql and Go installed. Search the best method fit for your operating system.

For Arch Linux, you can use pacman to install postgresql and go.
`sudo pacman -S postgresql go`

### Config

Manually create a config file in your home directory, ~/.gatorconfig.json, with the following content:

`{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"}`

### Installation

To install the gator CLI globally (so it can be run directly from any terminal session), you can use the go install command.

Ensure your Go environment (installed in the previous step) supports go install. Then, run the following command to install gator:

`go install github.com/alifoo/blog-aggregator@latest`

This will install the gator command from the module github.com/alifoo/blog-aggregator and make it accessible in production, rather than just development mode.

### Usage (commands)

Commands are executed via command-line using two different methods. For production (best way) you can type via terminal within anywhere in your machine:

`blog-aggregator <command> [arguments]`

For example, to register a user named "Alice":

`blog-aggregator register Alice`

For development mode, clone the repo and access the project root, then execute:

`go run . <command> [arguments]`

For example, to register a user named "Alice":

`go run . register Alice`

1. login
   `login <username>`

   _Logs in a user by setting their username in the configuration json._

2. register
   `register <username>`

   _Creates a new user and sets them as the current user._

3. reset
   `reset`

   _Deletes all users from the database._

4. users
   `users`

   _Lists all users in the system, marking the currently logged-in user._

5. agg
   `agg <interval>`

   _Fetches RSS feeds at a specified interval (e.g., "30s" for 30 seconds)._

6. addfeed
   `addfeed <feed_name> <feed_url>`

   _Adds a new RSS feed and follows it._

7. feeds
   `feeds`

   _Lists all available RSS feeds along with their owners._

8. follow
   `follow <feed_url>`

   _Follows an existing RSS feed._

9. following
   `following`

   _Lists all feeds the current user is following._

10. unfollow
    `unfollow <feed_url>`

    _Unfollows a feed._

11. browse
    `browse [limit]`

    _Displays the latest posts from followed feeds, with an optional limit (default: 2)._
