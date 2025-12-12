# rssgator

is a lightweight CLI tool to aggregate, manage and display content from multiplr RSS feeds.  Provides simple service to fetch post and commands for reading.

To run "gator" you will need to install:

-Go: version 1.20 or newer.

-PostgresSQL: Running Locally on your system.
__________________________________
== INSTALLATION ==

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/jaxhemopo/rssgator.git
    cd rssgator
    ```

2.  **Install the Binary:**
    ```bash
    go install .
    ```
__________________________________
== Set Environment Variable ==

`gator` requires the database connection string to be available as an environment variable named `DATABASE_URL`.

> **Example (You MUST update the username, password, and database name):**
>
> ```bash
> # This is the full, secure format. Replace the bracketed placeholders with your details.
> export DATABASE_URL="postgres://[your_user]:[your_password]@[host]:[port]/[db_name]"
> 
> # Example for a common local setup:
> # export DATABASE_URL="postgres://jaxhemopo:@localhost:5432/gator"
> ```

**Note on Local Setup:** If your PostgreSQL is configured for "trust" or "ident" authentication on your local machine, you may be able to omit the password, as shown in the example above (by leaving the password field blank but including the '@' symbol). Always use a secure password format for production deployments.

_______________________________________
== COMMANDS ==
* Account Management 
-----
% gator register <username>	
Creates a new user account in the database.
-----
% gator login <username>	
Sets the current CLI session to the specified user. Required for almost all other commands.
-----
% gator reset	
DANGER: Drops all database tables and resets the schema.

* Feed Management 
-----
% gator addfeed <name> <url>	
Adds a new RSS feed to the system and automatically follows it for the current user.
-----
% gator unfollow <url>
Stops following a feed. The feed remains in the system but posts will not show up in your feed.
-----
% gator following	
Displays a list of all feeds the current user is subscribed to.

* Aggregation
------
% gator agg <time_between_reqs>	
Starts the continuous aggregation service. This command runs forever (until interrupted with Ctrl+C). It fetches all feeds and saves new posts to the database based on the interval.
@ Example: gator agg 1m (Collects feeds every 1 minute)
-----
% gator browse [limit]	
Views the latest posts from all followed feeds, ordered by publication date.
@ Example: gator browse 10 (Shows the 10 most recent posts)



Example Workflow

After you have added your desired feeds with the (gator addfeed <name> <url>) command.

Start the Service (Terminal 1):

# This runs the aggregator in the background
gator agg 5s
Browse Content (Terminal 2):


# Switch to your reading terminal
gator login myuser
gator browse 5