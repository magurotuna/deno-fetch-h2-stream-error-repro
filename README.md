# deno-fetch-h2-stream-error-repro

This repository contains a minimal reproduciton for the HTTP2 stream error with Deno's `fetch`.

## Preparation

In this reproduction we use `nginx` as a HTTP2 server. Everything but DNS config is all set. To set up the DNS, add the following line to your `/etc/hosts` file:

```
127.0.0.1       deno-fetch-h2-repro.dev
```

Then run the following command to start up the `nginx` container:

```sh
$ docker compose up -d
```

If you want to confirm your environment is correctly set up, use `curl` command like below.

```sh
$ curl --insecure --http2 https://deno-fetch-h2-repro.dev
ok
```

## Reproduction

It contains programs written in Deno, Rust, and Go. Deno and Rust failed with the `REFUSED_STREAM` error while Go succeeded.

### Deno

```sh
$ cd deno
# You can change the concurrency
$ CONCURRENCY=5 deno task run
```

Output should look like:

```
DANGER: TLS certificate validation is disabled for all hostnames
❌ 2: TypeError: error sending request for url (https://deno-fetch-h2-repro.dev/): http2 error: stream error received: refused stream before processing any application logic
❌ 3: TypeError: error sending request for url (https://deno-fetch-h2-repro.dev/): http2 error: stream error received: refused stream before processing any application logic
❌ 4: TypeError: error sending request for url (https://deno-fetch-h2-repro.dev/): http2 error: stream error received: refused stream before processing any application logic
✅ 0: 200
✅ 1: 200
done
```

### Go

```sh
$ cd go
$ CONCURRENCY=5 go run main.go
✅ 2 ok
✅ 1 ok
✅ 0 ok
✅ 3 ok
✅ 4 ok
done
```

### Rust

```sh
$ cd rust
$ CONCURRENCY=5 cargo run
    Finished dev [unoptimized + debuginfo] target(s) in 0.23s
     Running `target/debug/rust`
❌ error sending request for url (https://deno-fetch-h2-repro.dev/): http2 error: stream error received: refused stream before processing any application logic

Caused by:
    0: http2 error: stream error received: refused stream before processing any application logic
    1: stream error received: refused stream before processing any application logic
✅ 1 ok
✅ 2 ok
❌ error sending request for url (https://deno-fetch-h2-repro.dev/): http2 error: stream error received: refused stream before processing any application logic

Caused by:
    0: http2 error: stream error received: refused stream before processing any application logic
    1: stream error received: refused stream before processing any application logic
❌ error sending request for url (https://deno-fetch-h2-repro.dev/): http2 error: stream error received: refused stream before processing any application logic

Caused by:
    0: http2 error: stream error received: refused stream before processing any application logic
    1: stream error received: refused stream before processing any application logic
done
```

## Root Cause

Enabling tracing in the Rust example helped to figure out the root cause.

According the actual log, it seems like the initial `SETTINGS` frame sent from the server did not reach the client when the client was initiating multiple concurrent streams. In the reproducible setting, the server (nginx) is configured to allow only 1 concurrent stream (`max_concurrent_streams = 1`) but this information reaches the client _after_ more streams have been started. This results in the server sending `REFUSED_STREAM` back to the client when the server receives more streams than is configured as per [RFC 9113 - 5.1.2. Stream Concurrency](https://datatracker.ietf.org/doc/html/rfc9113#name-stream-concurrency).

Note however that [RFC 9113 - 3.4. HTTP/2 Connection Preface](https://datatracker.ietf.org/doc/html/rfc9113#name-http-2-connection-preface) states that _this client behavior is totally valid_:

> To avoid unnecessary latency, clients are permitted to send additional frames to the server immediately after sending the client connection preface, without waiting to receive the server connection preface. It is important to note, however, that the server connection preface SETTINGS frame might include settings that necessarily alter how a client is expected to communicate with the server. Upon receiving the SETTINGS frame, the client is expected to honor any settings established.

## Possible Solution

One straightforward solution would be to __retry__ upon receipt of `REFUSED_STREAM`. [RFC 9113 - 8.7. Request Reliability](https://datatracker.ietf.org/doc/html/rfc9113#section-8.7) ensures that this type of error can be safely retried.

In Go, `REFUSED_STREAM` is treated as a retryable error.
https://cs.opensource.google/go/x/net/+/internal-branch.go1.21-vendor:http2/transport.go;l=655;bpv=1;bpt=0

libcurl does similarly.
https://github.com/curl/curl/blob/07bcae89d5d0090f1d11866d5f9c98c3720a2838/lib/http2.c#L1699-L1706
https://github.com/curl/curl/blob/07bcae89d5d0090f1d11866d5f9c98c3720a2838/lib/transfer.c#L1779-L1789

Also, a Rust crate `smithy-rs`, which powers `aws-sdk-rust`, has added its own retry logic on top of `hyper` when it receives `REFUSED_STREAM`.
https://github.com/smithy-lang/smithy-rs/pull/2971

## Validity of the Solution

I applied a quick patch that gets `REFUSED_STREAM` to be handled as a retryable error in `reqwest`.
https://github.com/magurotuna/reqwest/commit/bca765bdcff5adf568699d4d8ddef1b66d605b7a

You can try the patched version by running the following commands.

```sh
$ cd rust-patched
$ CONCURRENCY=5 cargo run
    Finished dev [unoptimized + debuginfo] target(s) in 0.11s
     Running `target/debug/rust`
✅ 0 ok
✅ 1 ok
✅ 2 ok
✅ 3 ok
✅ 4 ok
done
```
