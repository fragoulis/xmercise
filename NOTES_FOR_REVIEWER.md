The implementation is close what I would consider a production ready app.

It was written 95% by agents and curated based on my standards in the course of 5 hours split here and there, so it is only natural that some bad apples have made it through.

It might feel bloated for such a simple api (the core domain logic could be 100 lines of code if we wanted to be minimal) but the idea is not to be minimal, but to showcase how a real-world appplication could be structured. All in all, the exercise manifests the patterns that I would use to architect a service that is going to be part of a distributed application.

There are a bunch of topics that we could disucss further, namely:

* primary key generation for distributed systems
* rest api vs grpc
* api gateway, bff patterns
* authorization
* ddd/layering
* runtime deployment (vm, pod)/lifecycle
* database connection pooling
* database migration strategy
* logging, panicking, debugging, telemetry
* the outbox pattern and its caveats
* event publishing and consuming (order, idempotency)
