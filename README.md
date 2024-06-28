# TRIMM Redirects Traefik Middleware

## Introduction
This plugin is a Traefik middleware written in Go Lang.
The purpose of it is to redirect the user according to rules processed from the Central API backend.

## Traefik Plugin Configuration

### Static
#### traefik.yml

```yaml
experimental:
  plugins:
    redirects-traefik-middleware:
      moduleName: github.com/TRIMM/redirects-traefik-middleware
      version: "v0.2.0"
```

### Dynamic
#### http.yml

```yaml
http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        web:
        address: ":80"
      middlewares:
        - redirects-traefik-middleware

  services:
    service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000

  middlewares:
    redirects-traefik-middleware:
      plugin:
        redirects-traefik-middleware:
          redirectsAppURL: "redirects-app:8081"
```

## Service App Configuration

> **_NOTE:_**
> Make sure that the Central API is running as well. Since we need to synchronize data from there.

### Env variables

Copy the existing `.env.example` to a new `.env` and insert real values from your setup.

### Running the service

```bash
docker-compose up -d --build
```