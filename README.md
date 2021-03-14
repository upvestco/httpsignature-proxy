# httpsignature-proxy

tl;dr: Localhost HTTP Signatures proxy.

The Upvest API requires you to use [HTTP Signatures](https://tools.ietf.org/id/draft-ietf-httpbis-message-signatures-01.html) for extra layer of security to verify the call is coming from a real tenant as well as ensure the request hasn't been tampered with on the way.

This is good for security but can be cumbersome while developing. That's why this tool exists. You run it locally on your dev machine and use the localhost port in your Postman, Insomnia, etc. tool to make your calls to the Upvest Sandbox.

## Usage

```sh
$ ./httpsignature-proxy
HTTP Proxy to add HTTP Signatures to your requests.

Usage:
  httpsignature-proxy [command]

Available Commands:
  help        Help about any command
  start       Starts the proxy on localhost for signing HTTP-requests
  version     shows the application version

Flags:
      --config string   config file (default is $HOME/.httpsignature-proxy.yaml)
  -h, --help            help for httpsignature-proxy
  -v, --verbose         Verbose output

Use "httpsignature-proxy [command] --help" for more information about a command.
```

```sh
$ ./httpsignature-proxy start --help
```

## Configuration

You can use a config-file `.httpsignature-proxy` to collect your config without having to pass it in via command line arguments.


## Building locally

`goreleaser --snapshot --rm-dist`

## Building a new release

Add a new version tag in semver format and run goreleaser
