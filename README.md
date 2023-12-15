# Mail Feed

Converts emails into an RSS feed.

# Usage 
1. copy .env.sample to .env and fill in
2. `go run .`
3. `curl -X POST -H "Content-Type: application/json" -d '{"name": "My Feed"}' localhost:8080/inbox #create an inbox account`
4. Returned id is what will now be routed to `localhost:8080/rss/<id>` i.e all emails received on `<id>@domain.com` will be parsed and available on `localhost:8080/rss/<id>`