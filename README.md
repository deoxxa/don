# don

`don` is a read-only node that can be used as a Atom/Feed aggregator for the Ostatus network.
No account of your own is needed to use it, but you do need to have a publicly available
IP address with a domain name to get updates from the network.

[fknsrs.biz/p/don](https://www.fknsrs.biz/p/don)

I was interested by mastodon, but found it to be a little too heavyweight for
me. I want something I can run as a single binary with no other services. To
that end, I'm embarking on writing this - tentatively named don.

I'm not sure what it'll end up being. Right now it's ~~just an experiment in
plugging different protocols together~~ a rudimentary read-only client. Maybe
it'll be a full node implementation, but maybe not.

## Prebuilt Binaries

You can download a binary [from bintray](https://bintray.com/deoxxa/don/portable/dev).

* [macos/amd64](https://bintray.com/deoxxa/don/download_file?file_path=don_dev_darwin-amd64)
* [linux/amd64](https://bintray.com/deoxxa/don/download_file?file_path=don_dev_linux-amd64)
* [linux/arm](https://bintray.com/deoxxa/don/download_file?file_path=don_dev_linux-arm)
* [linux/arm64](https://bintray.com/deoxxa/don/download_file?file_path=don_dev_linux-arm64)
* [windows/amd64](https://bintray.com/deoxxa/don/download_file?file_path=don_dev_windows-amd64.exe)

On linux and macos you'll have to `chmod +x` the file after you download it.
Keep in mind that this is a terminal program, so if you double-click it or try
to open it from your downloads, you won't really see much happen. You'll need
to open a terminal (Terminal.app on macos, cmd.exe on Windows, xterm or
similar on linux), browse to the location of the binary, and run it like so:

```
$ ./don_dev_darwin-amd64 --public_url https://my-domain-name.com/
```

Change the filename and `public_url` argument to suit. Be sure to check out
the help output (via `--help`) to see the available options.

## Usage

```
usage: don --public_url=PUBLIC_URL [<flags>]

Really really small OStatus node.

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --addr=":5000"                 Address to listen on.
  --database="don.db"            Where to put the SQLite database.
  --public_url=PUBLIC_URL        URL to use for callbacks etc.
  --log_level=INFO               How much to log.
  --pubsub_refresh_interval=15m  PubSub subscription refresh interval.
  --record_documents             Record all XML documents for debugging.
```

All these options are available as environment variables as well - just make
them uppercase, e.g. `addr` is `ADDR`.

## Build Portable Binary

Right now, you'll need the following:

1. [go](https://golang.org/) (confirm via `go version`)
2. [docker](https://www.docker.com/) (confirm via `docker version`)

```
$ go get fknsrs.biz/p/don
$ cd $GOPATH/src/fknsrs.biz/p/don
$ make
```

You should see something like the following:

```
cd client && yarn run build-server
yarn run v0.16.1
$ NODE_ENV=production webpack -p --bail --config webpack.config.server.js
[long output omitted]
cd client && yarn run build-client
yarn run v0.16.1
$ NODE_ENV=production webpack -p --bail --config webpack.config.client.js
[long output omitted]
go build -ldflags=-s -o don
rice append --exec don
```

At the end, you'll have a self-contained binary named `don` that you can move
anywhere you like.

## Build Cross-Platform Binaries

This is the same as above, except that you'll need one more tool:

1. [xgo](https://github.com/karalabe/xgo)

Now, instead of running `make`, you run `make cross`. You'll end up with
binaries named `don-darwin-10.6-amd64`, `don-linux-amd64`, `don-linux-arm-5`,
and `don-windows-4.0-amd64.exe`.
