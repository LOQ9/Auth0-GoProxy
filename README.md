# Auth0 Go Proxy

**This repository is no longer maintained. Nowadays there are several zero trust alternatives such as Cloudflare Access, Pomerium, ...**

Sometimes, you want to put a sensitive web application behind a login wall.
Sometimes, you don't want to write the authentication logic yourself.
In this repository, we provide a proxy application that authenticates through Auth0.

## Required Parameters

The Auth0 Proxy requires a configured Auth0 client that is responsible for authenticating users.
Several of the parameters that the Auth0 Proxy module requires come from this client.

- `auth0_domain` - The domain you will use for Auth0 authentication.

- `auth0_client_id` - The ID of the Auth0 client to be used for authentication.

- `auth0_client_secret` - The secret string of the Auth0 client to be used for authentication.

- `auth0_redirect_uri` - The location to which Auth0 should redirect after authentication.
Unless you have a custom domain, this should be the URL of the Auth0 Proxy's load balancer.
**NOTE: This must contain the protocol, and must match a URL specified in the Auth0 client's allowed callback URLs.**

- `ssl_certificate_name` - The Auth0 Proxy communicates over https, so you must supply an SSL certificate.
For production services, it's strongly recommended that you create and use a different certificate.

This repository is based on https://github.com/quintilesims/auth0-proxy
