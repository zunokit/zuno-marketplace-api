# Tilt Configuration for New Services

When adding a new service, you need to register it in the `Tiltfile` located in the root directory.

## Template

Add the following block to the "Application Services" section of `Tiltfile`.
Replace `<service-name>` and `<port>` with your service details.

```python
# <Service Name>
build_service(
    '<service-name>',
    '<service-name>',     # Directory name in services/
    '<port>:<port>',      # Port forwarding (Host:Container)
    deps=['postgres', 'redis', 'rabbitmq'] # Add other dependencies if needed
)
```

## Verification

1. Run `tilt up` in the terminal.
2. Check the Tilt UI (http://localhost:10350) to ensure the service starts correctly.
