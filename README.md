# don

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

## Development

Most of the work will probably be in the client code. To make this easier,
there's a `live_reload` makefile target. This runs a couple of webpack
processes, and the don server, wiring the server up so that it uses the hot
reloading bits of webpack. If none of that made sense, don't worry. I've tried
to paper over all the details on purpose - configuring webpack is _not_ for
the faint of heart.

All you need to know is that when you run the live reloading environment, you
have to provide all your server parameters as environment variables. I suggest
something like the following:

```
$ export PUBLIC_URL=http://my-host.com
$ export LOG_LEVEL=DEBUG
$ make live_reload
```

Once it's all running, you should be able to open `http://127.0.0.1:5100/` in
your browser. You'll see a *very* quick flash of unstyled content, but then
the client JS should kick in and fix it up.

When this is running, you'll be able to save files and have the content in the
browser update automatically. This makes working on client stuff *much* nicer.

## Code Style

The main idea for the code style in this project is that it should be
automated. Not just automatically checked, but automatically applied. No
bikeshedding, no suggestions, no discussions. Computer is always right.

For go, use [gofmt](https://golang.org/cmd/gofmt/).

For JavaScript, use [prettier](https://github.com/prettier/prettier) with
`--single-quote` and `--trailing-comma es5`.

For CSS, use [csscomb](http://csscomb.com/) with the config provided in
`client/.csscomb.json`.

For shell scripts, use [shfmt](https://github.com/mvdan/sh) with `-i 2`.

## Acknowledgements

Some included icons were made by [Freepik](http://www.freepik.com) at
[Flaticon](http://www.flaticon.com), which were shared under the [CC 3.0
BY](http://creativecommons.org/licenses/by/3.0/) Creative Commons license.

The included username blacklist is based on [The Big Username
Blacklist](https://github.com/marteinn/The-Big-Username-Blacklist) by [Martin
Sandström](http://marteinn.se/), which was provided under the MIT license.

Some code for serialising forms was adapted from
[freiform](https://github.com/mechanoid/freiform) by [Falk
Hoppe](https://github.com/mechanoid), which was provided under the Apache
License 2.0.
