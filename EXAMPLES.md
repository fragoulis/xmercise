# Examples

```sh
curl -s localhost:8080/v1/companies/e29ab9c3-6795-4e42-843a-515427f65f32 -i
HTTP/1.1 404 Not Found
Content-Type: application/json
Date: Wed, 09 Sep 2026 08:15:49 GMT
Content-Length: 43

{"status":404,"title":"Company not found"}
```

```sh
❯ curl -s localhost:8080
{"status":404,"title":"Not found"}
❯ curl -s localhost:8080 -i
HTTP/1.1 404 Not Found
Content-Type: application/json
Date: Wed, 09 Sep 2026 08:17:29 GMT
Content-Length: 35

{"status":404,"title":"Not found"}
```

```sh
curl -s -XPOST localhost:8080/v1/companies -d'{"name":"foo", "employees_count":1, "type":"NonProfit"}' | jq
{
  "created_at": "2026-09-09T11:15:15.461177+03:00",
  "description": null,
  "employees_count": 1,
  "id": "c9f77fb9-5410-4086-a17d-9a90f31c132e",
  "name": "foo",
  "registered": false,
  "type": "NonProfit",
  "updated_at": "2026-09-09T11:15:15.461177+03:00"
}
```

```sh
curl -s localhost:8080/v1/companies/c9f77fb9-5410-4086-a17d-9a90f31c132e |jq
{
  "created_at": "2026-09-09T11:15:15.461177+03:00",
  "description": null,
  "employees_count": 1,
  "id": "c9f77fb9-5410-4086-a17d-9a90f31c132e",
  "name": "foo",
  "registered": false,
  "type": "NonProfit",
  "updated_at": "2026-09-09T11:15:15.461177+03:00"
}
```

```sh
curl -s -XPATCH localhost:8080/v1/companies/c9f77fb9-5410-4086-a17d-9a90f31c132e \
     -H 'Content-Type: application/merge-patch+json' \
     -d '{"description":"foo bar"}' | jq
{
  "created_at": "2026-09-09T11:15:15.461177+03:00",
  "description": "foo bar",
  "employees_count": 1,
  "id": "c9f77fb9-5410-4086-a17d-9a90f31c132e",
  "name": "foo",
  "registered": false,
  "type": "NonProfit",
  "updated_at": "2026-09-09T11:25:25.915338+03:00"
}
```

```sh
curl -s -XDELETE localhost:8080/v1/companies/c9f77fb9-5410-4086-a17d-9a90f31c132e
```
