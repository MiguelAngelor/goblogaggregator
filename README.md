# Gator (GUIDED PROJECT)

Gator is a command-line RSS feed aggregator written in Go. Users can register accounts, add RSS feeds, follow feeds, aggregate posts, and browse articles from the terminal.

## Requirements

Before running Gator, install:

- Go
- PostgreSQL

## Installation

Clone the repository:

```bash
git clone https://github.com/MiguelAngelor/goblogaggregator.git
cd goblogaggregator
```

Build the project:

```bash
go build
```

## Database Setup

Create a PostgreSQL database:

```bash
createdb gator
```

Run migrations:

```bash
cd sql/schema
goose postgres "postgres://YOUR_USERNAME:@localhost:5432/gator" up
```

## Configuration

Create a configuration file at:

```bash
~/.gatorconfig.json
```

Example:

```json
{
  "db_url": "postgres://YOUR_USERNAME:@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace `YOUR_USERNAME` with your PostgreSQL username.

## Commands

Register a user:

```bash
go run . register miguel
```

Login:

```bash
go run . login miguel
```

Add a feed:

```bash
go run . addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"
```

List all feeds:

```bash
go run . feeds
```

Follow a feed:

```bash
go run . follow "https://blog.boot.dev/index.xml"
```

Show followed feeds:

```bash
go run . following
```

Unfollow a feed:

```bash
go run . unfollow "https://blog.boot.dev/index.xml"
```

Aggregate posts:

```bash
go run . agg 1m
```

Browse posts:

```bash
go run . browse
```

## Features

- User registration and login
- RSS feed management
- Follow and unfollow feeds
- Automatic feed aggregation
- PostgreSQL storage
- Browse posts from followed feeds

## Development

Run the application during development:

```bash
go run . <command>
```

Build the executable:

```bash
go build
```
