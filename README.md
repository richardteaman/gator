# Gator CLI
Gator is a cli tool that aggregates RSS feeds and allows to browse posts from followed feeds.

## Requirements

Go version 1.23+  (https://go.dev/dl/)  
PostgreSQL (https://www.postgresql.org/download/)


## Installation

go install github.com/richardteaman/gator

## Configuration
Gator reads database credentials and your current user from a config file. 
The config file is: ~/.gatorconfig.json

currently .gatorconfig.json is not generated automatically (create with `nano` or similar)
example for it provided below:  
{  
    "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
    "current_user_name": "your-username"  
}

## Runnig the program
run with:  
**gator** [command] `<args>`

## Commands
**login** `<username>` -- swith current user  
**register** `<username>` -- create new user and login as it  
**reset** -- deletes everything to a blank state  
users -- lists all users and current  
**agg** `<period>`  -- starts aggregation with interval specified by user. When used without args default period is 2s. Period should look like; 10s, 2m , 3h   

**addfeed** `<feed name> <url>` -- adds feed to be aggregated later, also makes current user follow this feed      
 
**feeds** -- lists all feeds   
**follow** `<feed url> ` --  makes current user follow this feed  
**following** -- lists feed that current user follows  
**unfollow** `<url>`  -- unfollows feed for current user  
**browse** `<amount>` -- displays `amount`h latest posts. default amount is 2.