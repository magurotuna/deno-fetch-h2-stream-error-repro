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

It contains programs written in Deno, Go, and Rust. Only Deno failed with the `REFUSED_STREAM` error.

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
    Finished dev [unoptimized + debuginfo] target(s) in 0.04s
     Running `target/debug/rust`
✅ 1 ok
✅ 4 ok
✅ 0 ok
✅ 2 ok
✅ 3 ok
done
```
