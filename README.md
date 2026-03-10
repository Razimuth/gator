// README.md

Program:
Gator is a command line interpreter programmed in Go that is an RSS feed aggregator.
The program can
1. gator login - sets the current user in the config
    usage: login <username>
2. gator register - adds a new user to the database
    usage: register <username>
3. gator users - lists all the users in the database
    usage: users
4. gator addfeed - add RSS feeds from the internet and store in a postgresSQL database
    usage: addfeed <feedname> <feedUrl>
5. gator follow - follows RSS feeds from other users
    usage: follow <feedUrl>
6. gator unfolllow - unfollows RSS feeds
    usage: unfollow <feedUrl>
7. gator following - display feeds current user following
    usage: following
8. gator browse - browse the stored posts in the terminal
    usage: browse

Prerequisites
1. Go installed
There are two main installation methods
Option 1 (Linux/WSL/macOS): The Webi installer is the simplest way.
Just run this in your terminal:
    curl -sS https://webi.sh/golang | sh
Read the output of the command and follow any instructions.
Option 2 (any platform, including Windows/PowerShell): Use the official Golang installation instructions. On Windows, this means downloading and running a .msi installer package; the rest should be taken care of automatically.

2. PostgresSQL installed
It listens for requests on a port (Postgres' default is :5432), and responds to those requests. To interact with Postgres, first you will install the server and start it. Then, you can connect to it using a client like psql or PGAdmin.
Install Postgres v15 or later.
for macOS with brew
    brew install postgresql@15
for Linux / WSL (Debian). Here are the docs from Microsoft, but simply:
    sudo apt update
    sudo apt install postgresql postgresql-contrib
Ensure the installation worked. The psql command-line utility is the default client for Postgres. Use it to make sure you're on version 15+ of Postgres:
    psql --version
(Linux / WSL only) Update postgres password:
    sudo passwd postgres
Enter a password, and be sure you won't forget it. You can just use something easy like postgres.
Start the Postgres server in the background
for Mac: brew services
    start postgresql@15
for Linux:
    sudo service postgresql start
Connect to the server. I recommend simply using the psql client. It's the "default" client for Postgres, and it's a great way to interact with the database. While it's not as user-friendly as a GUI like PGAdmin, it's a great tool to be able to do at least basic operations with.
Enter the psql shell:
for Mac:
    psql postgres
for Linux:
    sudo -u postgres psql
You should see a new prompt that looks like this:
    postgres=#
Create a new database. I called mine gator:
    CREATE DATABASE gator;
Connect to the new database:
    \c gator
You should see a new prompt that looks like this:
    gator=#
Set the user password (Linux / WSL only)
    ALTER USER postgres PASSWORD 'postgres';
For simplicity, I used postgres as the password. Before, we altered the system user's password, now we're altering the database user's password.
Query the database
From here you can run SQL queries against the gator database. For example, to see the version of Postgres you're running, you can run:
    SELECT version();
You can type exit to leave the psql shell.

Other programs I installed to develop gator
1. Goose Migrations
Goose is a database migration tool written in Go. It runs migrations from a set of SQL files, making it a perfect fit for this project (we wanna stay close to the raw SQL).
Installing Goose
    go install github.com/pressly/goose/v3/cmd/goose@latest
Run goose -version to make sure it's installed correctly.
Create a users migration in a new sql/schema directory.
A "migration" in Goose is just a .sql file with some SQL queries and some special comments. 
I created a file in sql/schema called 001_users.sql with the following contents:
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL UNIQUE
);
-- +goose Down
DROP TABLE users;
The -- +goose Up and -- +goose Down comments are case sensitive and required. They tell goose how to run the migration in each direction.
I got my connection string. A connection string is just a URL with all of the information needed to connect to a database. The format is:
    protocol://username:password@host:port/database
for macOS (no password, your username):
     postgres://wagslane:@localhost:5432/gator
for Linux (password from last lesson, postgres user):
    postgres://postgres:postgres@localhost:5432/gator
Test your connection string by running psql, for example:
    psql "postgres://wagslane:@localhost:5432/gator"
It should connect you to the gator database directly.
Run the up migration to create the users table.
cd into the sql/schema directory and run:
    goose postgres "postgres://postgres:postgres@localhost:5432/gator" up

2. SQLC
SQLC is an amazing Go program that generates Go code from SQL queries.
I used Goose to manage our database migrations (the schema), and SQLC to generate Go code that our application can use to interact with the database (run queries).
Installing SQLC
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
Configure SQLC. You'll always run the sqlc command from the root of your project. Create a file called sqlc.yaml in the root of your project. Here is mine:
version: "2"
sql:
  - schema: "sql/schema"
    queries: "sql/queries"
    engine: "postgresql"
    gen:
      go:
        out: "internal/database"
We're telling SQLC to look in the sql/schema directory for our schema structure (which is the same set of files that Goose uses, but sqlc automatically ignores "down" migrations), and in the sql/queries directory for queries. We're also telling it to generate Go code in the internal/database directory.
Write a query to create a user. Inside the sql/queries directory, create a file called users.sql. Here's the format:
-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES ($1,$2,$3,$4)
RETURNING *;
Generate the Go code. Run sqlc generate from the root of your project. It should create a new package of go code in internal/database. You'll notice that the generated code relies on Google's uuid package, so you'll need to add that to your module:
    go get github.com/google/uuid
Import a PostgreSQL driver.
    go get github.com/lib/pq
Add this import to the top of your main.go file:
import _ "github.com/lib/pq"

You will need to create a config file
We'll use a single JSON file to keep track of two things:
Who is currently logged in
The connection credentials for the PostgreSQL database
The JSON file should have this structure (when prettified):
{
  "db_url": "connection_string_goes_here",
  "current_user_name": "username_goes_here"
}
Manually create a config file in your home directory, ~/.gatorconfig.json.  with the following content:
{
  "db_url": "postgres://example"
}
Don't worry about adding current_user_name, that will be set by the application.
Here is what mine looks like
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name": "Razimuth"
}

