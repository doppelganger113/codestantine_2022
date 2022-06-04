# Demo API

Main business logic service for Demo.

<div align="center">
    <img src="./assets/go_logo.png" align="center" width="200" alt="Go" />
</div>

Table of contents
=================

<!--ts-->

* [Configuration](#configuration)
* [Developing](#developing)
  * [Development requirements](#development-requirements)
  * [Running locally](#running-locally)
  * [Dependency management](#running-locally)
* [Open Api 3 documentation](#open-api-3-documentation)
* [Testing](#testing)
  * [Unit testing](#unit-tests)
  * [Integration testing](#integration-tests)
  * [Test configuration](#test-configuration)
* [Migrations](#migrations)
  * [Migration requirements](#migration-requirements)
  * [Running migrations](#running-migrations)
  * [Migrations in CI/CD](#migrations-in-cicd)
* [CI/CD](#cicd)
  * [Deploying to CI/CD](#deploying-to-cicd)
* [Troubleshooting](#troubleshooting)

<!--te-->

## Configuration

Configuration of the application is done through the environment variables, which are following:

**Required** environment variables:

| Environment variable name       | Required     | Default          | Description                                                                                                                                                                            |
|---------------------------------|--------------|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| PORT                            | Optional     | `3000`           | Http server port                                                                                                                                                                       |
| DEBUG_ROUTES                    | Optional     | `false`          | Default value is false, set to `true` for local endpoint debugging                                                                                                                     |
| AWS_REGION                      | **Required** |                  | Example: `eu-central-1`                                                                                                                                                                |
| AWS_USER_POOL_ID                | **Required** |                  | Example: `eu-central-1_somenumber`                                                                                                                                                     |
| AWS_ACCESS_KEY_ID               | **Required** |                  | AWS IAM key id used for API to talk with AWS services                                                                                                                                  |
| AWS_SECRET_ACCESS_KEY           | **Required** |                  | Secret for the above key                                                                                                                                                               |
| DATABASE_URL                    | **Required** |                  | Example `postgresql://postgres:example@localhost/demo?sslmode=disable`                                                                                                                 |
| IMAGES_API_DOMAIN               | **Required** |                  | Endpoint for the image service API                                                                                                                                                     |
| CORS_ALLOW_ORIGINS              | **Required** |                  | List of origins to allow CORS in format: `first.com, second.com, etc.com` or `http://localhost:4200`                                                                                   |
| SQS_POST_AUTH_URL               | **Required** |                  | Url of the SQS queue                                                                                                                                                                   |
| SQS_POST_AUTH_INTERVAL_SEC      | Optional     | `600`            | Interval in which the API will pool the queue for user registration events. Default value is `600`                                                                                     |
| SQS_POST_AUTH_CONSUMER_DISABLED | Optional     | `false`          | Set value to `true` to turn off in modes like local development to avoid messing with production                                                                                       |
| BASIC_AUTH_REALM                | Optional     | `Forbidden`      | Name of the realm for authentication                                                                                                                                                   |
| BASIC_AUTH_USERNAME             | Optional     |                  | Username used for basic authentication                                                                                                                                                 |
| BASIC_AUTH_PASSWORD             | Optional     |                  | Password used for basic authentication                                                                                                                                                 |
| OAUTH2_AUTHORIZATION_CODE_URL   | Optional     |                  | Url for OAuth2 authentication in format `https://twin-mirror.auth.eu-central-1.amazoncognito.com/login?response_type=code&client_id=<your-client-id>&redirect_uri=<your-redirect-uri>` |
| OAUTH2_TOKEN_URL                | Optional     |                  | Url for OAuth2 token retrieval in format `https://twin-mirror.auth.eu-central-1.amazoncognito.com/oauth2/token`                                                                        |
| DOMAIN                          | Optional     | `localhost:3000` | Name of the domain the app is being served from, like `localhost:3000` or `https://twin-mirror.herokuapp.com`                                                                          |

## Developing

### Development requirements

- Go v1.18+
- PostgreSQL
- Docker and docker-compose

You will need to create a new aws cli profile locally and use it with credentials for this API. To create it
execute `aws configure --profile your-profile` then `export AWS_PROFILE=your-profile`

### Running locally

1. Set the environment variables. It's also recommended storing them in `.env` file which is ignored for ease of
   management, copy the example `cp .env.example .env`. Then load the env variables with `export $(cat .env | xargs)`
2. Ensure that you start the database and all the other required services by running:
   `docker-compose up -d`
3. After setting the environment variables, execute the `make start` command to build and start the server

### Dependency management

**Remove dependency** by removing all occurrences of the library in imports and execute:

```bash
go mod tidy
```

### Dependency injection

Dependency injection is done with with Google's [Wire](https://github.com/google/wire) that does build time dependency
management. Construct dependencies in `wire.go` and building is done with the following command:

```bash
wire
```

## Open Api 3 documentation

Golang implementation of [OpenApi3 specification](https://swagger.io/docs/specification/basic-structure/) aka Swagger
through dynamic configuration with Swagger UI.

We use Swagger-UI with some small changes in order for it to fetch changes from our API where we can set the
redirection url for OAuth2.

**Dependencies**

- github.com/getkin/kin-openapi/openapi3
- github.com/getkin/kin-openapi/openapi3gen

## Testing

### Unit testing

- `make test`

## Migrations

### Migration requirements

To run the migrations, you must set the database connection url through environment variable:

```bash
export DATABASE_URL=postgresql://postgres:example@localhost/demo?sslmode=disable
```

Note that the `sslmode=disable` is for local development only, for production you should use ssl encryption.

### Running migrations

**BE CAREFUL WHEN REVERTING!**
Migrations are done with Go library [golang-migrate/migrate](https://github.com/golang-migrate/migrate) and are executed
with following commands:

- `make migrate_up` to update the schema to the latest
- `make migrate_up_step` to update the migration one step up
- `make migrate_down` to revert the schema one step down
- `go run ./cmd/migrations/main.go -f 3` force migration to specified version

It's paramount to follow [best practices](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) to ensure
you don't break anything.

### Migrations in CI/CD

For CI/CD ensure that you always run migrations before deploying your new app, like a pre-run action. For rollbacks,
also ensure that you first close the application before running the migration down command. As always, for specific
cases you will need to pay attention if you need code that supports both versions, so it doesn't crash if the change is
drastic.

## CI/CD

CI/CD is currently on the Heroku and additional options that were added for it are located in `go.mod` file as:

1. Build phase

```text
// +heroku goVersion 1.18
// +heroku install ./cmd/...
```

2. Run phase in `Procfile`

They specify the Go version and build location. Builds will end up in `bin/` directory. Other thing to note is that
the `bin/go-post-compile` and `bin/go-pre-compile` will execute if they exist, so use them for pre and post actions.

### Deploying to CI/CD

It is recommended to execute both unit and integration tests, as well as building the app before pushing to git.

```bash
# Test
export TEST_INTEGRATION=true
make test
make
```

More about Heroku options can be found at the
[Go Buildpack](https://github.com/heroku/heroku-buildpack-go#prepost-compile-hooks).

## Troubleshooting

- Error when building: `open /usr/local/go/pkg/darwin_amd64/runtime/cgo.a: permission denied`
   ```bash
   sudo chmod -R 777 /usr/local/go
   ```
- New to Makefile? There's plenty of learning examples [here](https://makefiletutorial.com/)
- How to clear test cache? Execute `go clean -testcache`
