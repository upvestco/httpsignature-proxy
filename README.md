# httpsignature-proxy

Localhost HTTP Signatures proxy and webhook events listener.

The Upvest Investment API requires you to
use [HTTP Signatures](https://tools.ietf.org/id/draft-ietf-httpbis-message-signatures-01.html)
for an extra layer of security to verify the call is coming from a real tenant as
well as ensure the request hasn't been tampered with on the way.

This is good for security but can be cumbersome while developing. That's why
this tool exists. You run it locally on your dev machine and use the localhost
port in your Postman, Insomnia, etc. tool to make your calls to the Upvest
Sandbox amd be able to see incoming webhook events.

## Installation

You can download the binaries from
the [Releases-page](https://github.com/upvestco/httpsignature-proxy/releases)

OR

You can install it from Homebrew:

```shell
brew tap upvestco/httpsignature-proxy
brew install httpsignature-proxy
```

## Building locally

```shell
git clone https://github.com/upvestco/httpsignature-proxy.git
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
  -p, --port int                      port to start server
  -v  --verbose-mode bool             enables verbose mode for proxy (not recommended to use with -l flag)
  -f, --private-key string            filename of the private key file
  -P, --private-key-password string   password of the private key
  -s, --server-base-url string        server base URL to pipe the requests to
  -i, --key-id string                 key id for specified private key
  -c, --client-id string              client id for specified private key
  -l, --listen                        create webhook events tunel and log incomming events to console
  -e  --events strings                only with -l flag. Log only events of the specified types
      --show-webhook-headers          only with -l flag. Show http headers coming with webhook events

Global Flags:
      --config string   config file (default is $HOME/.httpsignature-proxy.yaml)
```

## Key generation

Upvest Investment API supports ECDSA and ed25519 types of private/public key
pair.

## Generate ECDSA key pair

To generate private key which can be used with http proxy use this command:

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

```shell
openssl ec -in ./ec-encr-priv-key.pem -pubout > ec-pub-key.pem
```

Generated key should be in PEM format. You can see an example in
`private_key_example.ppk` (password:`123456`)

**Please note that the httpsignature-proxy is designed to use ECDSA key only.**

## Generate ed25519 key pair

OSX does not support the native generation of ed25519 private/public key pair.
You can use this way of generation **only on OS Unix based systems**.

Generate private ed25519 key:

```sh
openssl genpkey -algorithm ed25519 -outform PEM -out ed25519.pem
```

Extract public key from private:

```sh
openssl pkey -outform DER -pubout -in ed25519.pem | tail -c +13 | \
openssl base64 > ed25519.pub
```

**Please note that the httpsignature-proxy does not support es25519 key type.
Despite the fact that httpsignature-proxy supports not protected by password
private keys, we strongly recommend to use only keys with password.**

## Configuration

You can configure your proxy in a few different ways:

- Passing in all config as command-line arguments
- Specifying a configuration file to use
- Exposing the config in environment variables

### Config-file

You can use a config-file `.httpsignature-proxy.yaml` to collect your config
without having to pass it in via command line arguments. Config-file support
more than one private key.

Please see `.httpsignature-proxy.sample` for reference.

## Example of usage

You can do a test request with the sample config. To do it you should:

- Rename `.httpsignature-proxy.sample` to `httpsignature-proxy.yaml`
- Start signature proxy:

```sh
./httpsignature-proxy --config ./httpsignature-proxy.yaml start
```

- Do a request:

```sh
curl -X GET "http://localhost:3000/headers" -H "accept: application/json"
```

## Authors

- [Kiryl Yalovik](https://github.com/kiryalovik)
- [Mike Konobeevskiy](https://github.com/upvest-mike)
- [Juha Ristolainen](https://github.com/upvest-juha)
