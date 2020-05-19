dp-observation-api
================
dp-observation-api

### Getting started

* Run `make debug`

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default                | Description
| ---------------------------- | ---------------------- | -----------
| BIND_ADDR                    | :24500                 | The host and port to bind to
| SERVICE_AUTH_TOKEN           | ""                     | The token used to identify this service when authenticating
| DATASET_API_URL              | http://localhost:22000 | The host name for the dataset API
| OBSERVATION_API_URL          | http://localhost:24500 | The host name for the observation API
| ZEBEDEE_URL                  | http://localhost:8082  | The host name for Zebedee
| DEFAULT_OBSERVATION_LIMIT    | 1000                   | The default limit number of observations returned in a reauest
| ENABLE_PRIVATE_ENDPOINTS     | false                  | Flag to enable private endpoints for the API
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                     | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                    | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                    | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

