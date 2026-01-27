# Apolon Demo

A simple demo application to test Apolon.

## Setup

### 1. Start PostgreSQL

```bash
cd docker
docker compose up -d
```

### Generate Field Accessors

Before running the demo, generate the type-safe field accessors:

```bash
# From the project root directory
go run ./apolon-cli generate -i ./apolon-demo -o ./apolon-demo
```

Or using `go generate` (from project root):

```bash
go generate ./apolon-demo/...
```

This creates smth like a "partial class" `model_fields.go` from `model.go`,
providing typed field accessors like `PatientFields.Age.Gt(18)`.

### 3. Run the Demo

```bash
go run .
```

## How to use: ##
1. Define your model struct in `model.go`
2. Run `apolon generate` to create `model_fields.go`
3. Use generated fields in your queries: `PatientFields.Age.Gt(18)`
4. If you change the model, regenerate the fields

You can automate this with `go:generate`:

```go
//go:generate go run github.com/jkeresman01/apolon/apolon-cli generate -i . -o .

type Patient struct {
    ID   int    `apolon:"id,pk"`
    Name string `apolon:"name"`
    Age  int    `apolon:"age"`
}
```

Then run `go generate ./...` from the project root to regenerate all fields.

Field accessor are already generated for this demo example so you can just use 

```bash
go run .
```
