# httpsignature-proxy

tl;dr: Localhost HTTP Signatures proxy.

The Upvest Investment API requires you to
use [HTTP Signatures](https://tools.ietf.org/id/draft-ietf-httpbis-message-signatures-01.html)
for extra layer of security to verify the call is coming from a real
tenant as well as ensure the request hasn't been
tampered with on the way.

This is good for security but can be cumbersome while developing.
That's why this tool exists.You run it locally on your
dev machine and use the localhost port in your Postman, Insomnia, etc.
tool to make your calls to the Upvest Sandbox.

## Installation

You can download the binaries from the [Releases-page](https://github.com/upvestco/httpsignature-proxy/releases)

OR

You can install it from Homebrew:

```shell
brew tap upvestco/httpsignature-proxy
brew install httpsignature-proxy
```

## Building locally on macOS

```shell
make
```

## Usage

```sh
$ ./httpsignature-proxy start --help
Starts the proxy on localhost for signing HTTP-requests

Usage:
  httpsignature-proxy start [flags]

Flags:
  -h, --help                          help for start
  -f, --private-key string            filename of the private key file
  -P, --private-key-password string   password of the private key
  -s, --server-base-url string        server base URL to pipe the requests to
  -i, --key-id string                 key id for specified private key
  -p, --port int                      port to start server

Global Flags:
      --config string   config file (default is $HOME/.httpsignature-proxy.yaml)

```

## Key generation

To generate private key which can be :used with http proxy use this command:

```sh
openssl ecparam -name prime256v1 -genkey -noout -out ./ec-priv-key.pem
```

After that you need to encrypt your key with the password:

```sh
openssl ec -in ./ec-priv-key.pem -out ./ec-encr-priv-key.pem -aes256
```

Remove unused key:

```sh
rm ./ec-priv-key.pem

```

Extract public key from private:

```
openssl ec -in ./ec-encr-priv-key.pem -pubout > ec-pub-key.pem
```

Generated key should be im PEM format. You can see example in
`private_key_example.ppk` (password:`123456`)

## Configuration

You can configure your proxy in a few different ways:

- Passing in all config as command-line arguments
- Specifying a config-file to use
- Exposing the config in environment variables

### Config-file

You can use a config-file `.httpsignature-proxy.yaml` to collect your config
without having to pass it in via command line arguments.

Please see `.httpsignature-proxy.sample` for reference.

### Environment variables

You can use environment variables to collect you config without having
to pass it in via command line arguments

```sh
export HTTP_PROXY_PRIVATE_KEY="/absolute/path/to/your_key.ppk"
export HTTP_PROXY_PRIVATE_KEY_PASSWORD="secret"
export HTTP_PROXY_SERVER_BASE_URL="https://someurl"
export HTTP_PROXY_KEY_ID="your key id"
export HTTP_PROXY_PORT=3000
```

## Example of usage

You can do test request with the sample config. To do it you should:

- Rename `.httpsignature-proxy.sample` to `.httpsignature-proxy.yaml`
- Start signature proxy:

```sh
./http-signature-proxy --config .http.signature-proxy.yaml
```

- Do some request:

```sh
curl -X GET "http://localhost:3000/headers" -H "accept: application/json"
```

## Authors

- [Kiryl Yalovik](https://github.com/kiryalovik)
- [Juha Ristolainen](https://github.com/upvest-juha)
