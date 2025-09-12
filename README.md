# URL Shortener

A simple and fast web application for shortening long URLs.  
Built with **Go** and containerized using **Docker Compose**.

### 🔗 [Link to app page](https://url-short-af.space) (currently unavailable)

## Tech Stack

- **Go** — backend implementation
- **PostgreSQL** — primary data storage
- **Redis** — caching / rate limits (if enabled)
- **ClickHouse** — analytics storage for redirect stats (if enabled)
- **Prometheus + Grafana** — metrics and dashboards (if enabled)
- **Nginx** — reverse proxy and TLS termination (if enabled)
- **Docker Compose** — container orchestration

## API Documentation

The API is described using Swagger (`swagger.json`).  
Main available endpoints:

### 🔗 Shorten URL

**POST** `/shorten`  
Receives a long link and returns a shortened one.  
Request body:

```json
{
  "link": "https://google.com"
}
```

Response example:

```json
{
  "link": "http://host/12345678"
}
```

---

### ↪️ Redirect by Code

**GET** `/r/{short_code}`  
Redirects to the original link by its short code with `308 Permanent Redirect`.

Path parameter:
- `short_code` — short identifier of the link

Responses:
- **308** — Successfully redirected
- **404** — Invalid code or internal error

---

### 📊 Link Statistics

**GET** `/stats/{short_code}`  
Returns statistics about link usage.

Path parameter:
- `short_code` — short identifier of the link

Response example:

```json
{
  "code": "abcd1234",
  "link": "https://google.com",
  "clicks": 100500
}
```

---

## License

This project is open-source and available under the MIT License.
