# don

[fknsrs.biz/p/don](https://www.fknsrs.biz/p/don)

I was interested by mastodon, but found it to be a little too heavyweight for
me. I want something I can run as a single binary with no other services. To
that end, I'm embarking on writing this - tentatively named don.

i'm not sure what it'll end up being. right now it's ~~just an experiment in
plugging different protocols together~~ a rudimentary read-only client. maybe
it'll be a full node implementation, but maybe not.

## build instructions

These assume that you have Go already running.

```shell
go get fknsrs.biz/p/don
go build fknsrs.biz/p/don
go install fknsrs.biz/p/don
don --help
```
