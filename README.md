# Stratus

Stratus is a command-line tool for publishing protobuf-based [cloudevents](https://cloudevents.io/)
onto a Kafka (or Kafka-compatible, e.g Redpanda) topic.

It's a tool primarily intended for testing microservices that rely on
events from a message broker, allowing you to avoid writing custom seed
commands.

## Status

Stratus is currently in development and there is no build or release process.
If you want to try it, clone it down, and run `make build`.
