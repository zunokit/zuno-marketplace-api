# Build Scripts for New Services

When adding a new service, create a `<service-name>-build.bat` in this directory to support manual local building.

## Template

Replace `<service-name>` with your service name.

```batch
@echo off
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o build/<service-name> ./services/<service-name>/cmd
```
