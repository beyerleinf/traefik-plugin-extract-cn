# Extract Common Name Traefik Plugin

This plugin can - when used in conjunction with the built-in `PassTLSClientCert` Middleware - extract the Common Name and put it into an HTTP Header with the defined name. This plugin requires you to enable the cert info on the `PassTLSClientCert` Middleware with at least the Common Name enabled.

```yml
experimental:
  plugins:
    extractCommonName:
      modulename: "github.com/beyerleinf/traefik-plugin-extract-cn"
      version: "v1.0.0"
```

### Dynamic

```yml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares:
        - "passClientCert
        - "extractCommonName"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    passClientCert:
      passTLSClientCert:
        pem: false
        info:
          subject:
            commonName: true
    extractCommonName:
      plugin:
        extractCommonName:
          destHeader: "X-Target"
```
