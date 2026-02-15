# Tools

Development tools for Zuno Marketplace API.

## Service Generator

Tool để tạo microservice mới với cấu trúc chuẩn và các file boilerplate.

### Cách Sử dụng

#### Option 1: Từ Root Directory (Recommended)

```bash
# Từ thư mục root của project
make generate NAME=payment
make generate NAME=notification
make generate NAME=analytics
```

#### Option 2: Windows Batch Script

```cmd
cd tools
create.bat payment
create.bat notification
```

#### Option 3: PowerShell Script

```powershell
cd tools
.\create.ps1 payment
.\create.ps1 notification
```

#### Option 4: Go Run Directly

```bash
cd tools
go run create_service.go -name=payment
```

### Files Generated

Khi chạy tool, sẽ tạo structure như sau:

```
services/{name}-service/
├── cmd/
│   └── main.go              # gRPC server entry point
├── internal/
│   ├── config/
│   │   ├── config.go        # Configuration loader
│   │   └── errors.go        # Config errors
│   ├── server/
│   │   └── server.go        # gRPC handlers
│   ├── service/
│   │   └── service.go       # Business logic
│   ├── repository/          # Data access layer
│   └── models/              # Domain models
├── .air.toml                # Hot reload config
└── README.md                # Complete setup guide
```

### Next Steps After Generation

Sau khi tạo service, làm theo hướng dẫn trong `services/{name}-service/README.md`:

1. **Add Environment Variables** - Thêm vào root `.env.development`:

   ```bash
   {NAME}_GRPC_PORT=:4XXX
   {NAME}_SERVICE_URL=localhost:4XXX
   ```

2. **Add Makefile Build Command** - Thêm vào root `Makefile`:

   ```makefile
   build-{name}: ## Build {name} service
       @echo Building {name}-service...
       @$(MKDIR_CMD)
       cd services/{name}-service/cmd && go build $(LDFLAGS) -o ../../../build/{name}-service$(if $(filter $(OS),Windows_NT),.exe,) .
   ```

3. **Create Proto File** (optional) - Tạo `proto/{name}.proto`

4. **Update Dev Script** - Thêm vào `scripts/dev.sh`

5. **Start Development**:
   ```bash
   ./scripts/dev.sh {name}
   ```

## Build the Tool

Để build tool thành executable:

```bash
cd tools
go build -o create_service.exe create_service.go
```

Sau đó có thể chạy trực tiếp:

```bash
./create_service.exe -name=payment
```

## Troubleshooting

### Error: "service already exists"

Service folder đã tồn tại. Xóa folder cũ trước:

```bash
rm -rf services/{name}-service
```

### Error: "go: command not found"

Cài đặt Go: https://go.dev/dl/

### Make command không hoạt động trên Windows

Dùng PowerShell script thay thế:

```powershell
cd tools
.\create.ps1 payment
```

## Contributing

Khi cập nhật tool:

1. Test với nhiều service names khác nhau
2. Verify generated files compile successfully
3. Check README templates có chính xác không
4. Test trên cả Windows và Unix-like systems
