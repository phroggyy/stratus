# Stratus

![Screenshot of Stratus](https://raw.githubusercontent.com/phroggyy/stratus/main/screenshot.png)

Stratus is a command-line tool for publishing protobuf-based [cloudevents](https://cloudevents.io/)
onto a Kafka (or Kafka-compatible, e.g Redpanda) topic.

It's a tool primarily intended for testing microservices that rely on
events from a message broker, allowing you to avoid writing custom seed
commands.

## Installation

The latest release can always be found under the [Releases](https://github.com/phroggyy/stratus/releases)
page.

Download the latest archive, unarchive it, and move the binary to a directory in your `$PATH`,
such as `/usr/local/bin`.

## Status

This project is a proof-of-concept build, with a lot of things not working.

Contributions are very welcome to making this more stable.

## To-do

Here's a list of things that aren't working properly

- [ ] Handle local imports – this is hard because we don't necessarily know where the root is (if we depend on Buf, we could traverse up until we find a `buf.gen.yaml`)
- [ ] Handle imports from other projects - when using `protoc`, this is the same problem as handling local imports. However, with Buf, these dependencies are kept remotely
      and it would be great to support Buf, as `protoc` is painful to work with when using packages.
- [ ] Support changing the connection details for the kafka broker – right now it always defaults to `localhost:9092`
- [ ] Better error handling when kafka connection fails – we should give the user an error modal and make them change the connection details if they try to hit send
      while the connection is broken.

