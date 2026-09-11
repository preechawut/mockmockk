# mockmockk
> Define HTTP method, path, and custom JSON response instantly

---

## Tech Stack

- Go, Fiber, Postgres, Docker Compose

---

## API Usage

```bash
# create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "E-Commerce", "slug": "shop"}'

# create mock endpoint in project
curl -X POST http://localhost:8080/api/v1/projects/<PROJECT_ID>/endpoints \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/products",
    "status": 200,
    "response": [{"id": 1, "name": "Keyboard"}]
  }'

curl http://localhost:8080/mock/shop/products
```