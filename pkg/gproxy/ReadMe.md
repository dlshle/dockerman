# gproxy
gproxy is a simple proxy *server* written in Go.

## capabilities
- TCP proxy
- backend load balancing with custom policies
- HTTP endpoint for proxy managements

## Proxy Management API
- GET /cfg - get current proxy configuration
- PUT /cfg - set new proxy configuration

sample configuration:
```json
{
  "upstreams": [
    {
      "port": 123,
      "backends": [
        {
          "host": "127.0.0.1",
          "port": 456
        }
      ],
      "policy": "roundrobin",
      "protocol": "tcp"
    }
  ]
}
```