# Starting the app
`make start`

After app is running, in a new terminal run: `make createEvents`

This will spin up the event stream and start sending events to the app.

# Scale
Currently the limiting factor is the database. With a migration we could add an index to the domain for faster lookups,
we could add in a redis store as well for a caching layer so that we dont need to query the DB each time, also we could
use something like pgbouncer to scale the database horizontally, adding in a replica could lighten the load for reads
(Not needed if we use redis).

Currently running off my system running with a database it takes about 2.3 seconds to process 10,000 events. If I switch
it over to use the memory adapter it takes 485ms to process 10,000 events.

![image.png](.docs%2Fimage.png)

# Design decisions
The system is a bit overengineered for the task at hand, but one of the statements in the task 
was that the system should be production ready, so adding in logging, error handling, and a way
to easily add/remove services was a priority. With this we can easily add new features, different
databases, extend the middleware, etc.

The file structure is so you can have a very easy mental model of the application as it grows
in size. 
* App - this is where the main application code lives.
  * cmd - this is where and extra commands live, like the event stream. 
  * Handlers - The handlers are the entry points for the application. They are responsible for
    * Middleware
    * Validating the request
    * Calling the service
    * Returning the response
  * main.go - this is the entry point for the application. It is responsible for
    * Setting up the application
    * Starting the application
* business - this is where the business logic lives. It is responsible for
  * models - this is where the models live
  * adapters - These are the concrete implementations of the interfaces. They are responsible
    for
    * Calling the database
    * Calling the external services
  * ports - These are the interfaces that the adapters implement. They are responsible for
    * Defining the methods that the adapters need to implement
    * Defining the models that the adapters need to use
* foundation - this is where reusable code would live, this code can/should live outside this project in its own repo

# Improvements
* Add in a redis cache to speed up the lookups
* Add in a pgbouncer to scale the database horizontally
  * Or just put it behind amazons RDS, allows for easy vertical scaling if needed and pretty straight forward horizontal 
  scaling with read replicas or by adding in a proxy (amazon rds proxy)
* Add in a replica to lighten the load on the database
* Add in a migration to add an index to the domain column
* Add in more tests for potential edge cases
  * Tests are first class citizens in my opinion, code is not ready for production until it has adequate tests. Meaning
  that the tests should cover the entire expected behavior of the code. This includes edge cases, and error handling.
* Add opentelemetry to add in tracing
* Add prometheus to add in metrics
* Add in a health check endpoint
* Add in more middleware to handle panics, metrics, Auth, etc
* Setup a CI/CD pipeline
* Setup grafana to monitor the application
* Setup pulumi(terraform) to manage the infrastructure
* Complete the dockerfile for production

# Testing
`make test`

# Tools
* [air](github.com/cosmtrek/air) - This is a live reload tool for go. It is used to speed up development.
* [make](https://www.gnu.org/software/make/) - This is a tool for running commands. It used to make commands reusable.
* [docker](https://www.docker.com/) - This is a tool for running containers.  Allows us to run the application in a
    consistent environment.
* [docker-compose](https://docs.docker.com/compose/) - This is a tool for running multiple containers that are related.
