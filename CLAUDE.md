# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Pulumi component provider written in Go that provides reusable AWS infrastructure components for web hosting. It uses the `pulumi-go-provider` framework with the `infer` package to automatically generate provider schemas from Go types and functions.

The provider is designed for **Git-based distribution** (not published to the Pulumi Registry), following Pulumi's recommended pattern for Git-hosted component providers.

## Build and Development Commands

```bash
# Build the provider binary
make build                    # Builds to ./bin/pulumi-resource-web-components
go build -o bin/pulumi-resource-web-components .

# Install provider locally for testing
make install                  # Installs to ~/.pulumi/plugins/resource-web-components-v0.1.0/

# Build and test
go build ./...               # Build all packages
go test ./...                # Run all tests
make test                    # Same as above

# Code quality
make fmt                     # Format all Go files
make lint                    # Run golangci-lint (requires installation)
go mod tidy                  # Clean up dependencies

# Clean artifacts
make clean                   # Remove build artifacts and installed provider
```

## Architecture and Code Structure

### Current Structure (Git-based Provider)

```
web-components-go/
├── main.go                          # Provider entry point (root level)
├── pkg/                             # Component implementations
│   ├── acm/certificate.go           # DnsValidatedCertificate component
│   └── web/environment.go           # WebEnvironment component
├── go.mod                           # Module: github.com/pulumi-initech/web-components-go
├── Makefile                         # Build automation
├── PulumiPlugin.yaml               # Plugin metadata (runtime: go)
└── schema.json                      # Provider schema (51718 bytes)
```

**Important**: The structure uses `pkg/` at the root level (not `provider/pkg/`). The main.go imports from `github.com/pulumi-initech/web-components-go/pkg/acm` and `pkg/web`.

### Component Registration Pattern

Components are registered in `main.go` using the `infer.ComponentF` wrapper:

```go
provider, err := infer.NewProviderBuilder().
    WithComponents(
        infer.ComponentF(acm.NewDnsValidatedCertificate),
        infer.ComponentF(web.NewWebEnvironment),
    ).
    Build()
```

### Component Implementation Pattern

Each component follows this structure:

1. **Args struct**: Defines component inputs with `pulumi` struct tags
   ```go
   type FooArgs struct {
       Bar pulumi.StringInput `pulumi:"bar"`
       Baz pulumi.IntPtrInput `pulumi:"baz,optional"`
   }
   ```

2. **Component struct**: Embeds `pulumi.ResourceState` and the Args struct
   ```go
   type Foo struct {
       pulumi.ResourceState
       FooArgs
       OutputField pulumi.StringOutput `pulumi:"outputField"`
   }
   ```

3. **Constructor**: Creates and registers the component
   ```go
   func NewFoo(ctx *pulumi.Context, name string, args FooArgs, opts ...pulumi.ResourceOption) (*Foo, error) {
       comp := &Foo{}
       err := ctx.RegisterComponentResource("web-components:module:Foo", name, comp, opts...)
       // ... create child resources with pulumi.Parent(comp) ...
       return comp, nil
   }
   ```

4. **Annotate method**: Provides schema descriptions (uses `infer.Annotator`)
   ```go
   func (f *Foo) Annotate(a infer.Annotator) {
       a.Describe(&f, "Component description")
       a.Describe(&f.Bar, "Bar input description")
   }
   ```

### Type Conversion Patterns

**Critical conversions when working with Pulumi types:**

- `pulumi.StringArray` → `pulumi.ArrayInput`: Use `pulumi.Array{...}.ToArrayOutput()`
  ```go
  AlarmActions: pulumi.Array{scaleOutPolicy.Arn}.ToArrayOutput()
  ```

- `pulumi.StringPtrInput` → `pulumi.StringInput`: Use `.ToStringPtrOutput().Elem()`
  ```go
  ZoneId: args.ZoneId.ToStringPtrOutput().Elem()
  ```

- Component resource registration: Use explicit type tokens (not `p.GetTypeToken(ctx)`)
  ```go
  ctx.RegisterComponentResource("web-components:module:ComponentName", name, comp, opts...)
  ```

### Provider Version Management

- Version is defined in `main.go` as `providerVersion` constant
- Current version: `0.1.2`
- Makefile uses `VERSION ?= 0.1.0` (may be out of sync with main.go)

## Component Details

### DnsValidatedCertificate (pkg/acm/certificate.go)

Creates an ACM certificate with automatic DNS validation via Route53. Implements a common pattern where:
1. ACM certificate is created with DNS validation
2. Route53 record is created for validation
3. CertificateValidation resource waits for validation to complete

**Type token**: `web-components:acm:DnsValidatedCertificate`

### WebEnvironment (pkg/web/environment.go)

Creates a complete auto-scaling web hosting environment with:
- Auto Scaling Group with Launch Template (Nginx installed via user data)
- Application Load Balancer with HTTPS listener and HTTP→HTTPS redirect
- Security groups for ALB and instances
- TLS-generated SSH key pair
- CloudWatch alarms for request-based scaling (high/low thresholds)
- Optional Route53 alias record

**Type token**: `web-components:web:WebEnvironment`

**Key implementation details**:
- Uses `base64.StdEncoding.EncodeToString()` for EC2 user data
- CloudWatch alarms trigger scaling policies based on ALB RequestCount metric
- Route53 alias record creation is conditional on both `zoneId` and `subdomain` being provided
- ALB has both HTTPS (port 443) and HTTP (port 80) listeners

## Common Issues and Solutions

### Build Errors

1. **"undefined: p.Annotator"**: Use `infer.Annotator` not `p.Annotator` (from `pulumi-go-provider/infer`)

2. **"cannot use pulumi.StringArray as pulumi.ArrayInput"**: Convert with `pulumi.Array{...}.ToArrayOutput()`

3. **"cannot use StringPtrInput as StringInput"**: Convert with `.ToStringPtrOutput().Elem()`

4. **Unused import "fmt"**: Remove unused imports (linter will catch these)

### Testing the Provider

After `make install`, the provider is available for local testing. Reference it in a Pulumi program using the Git URL pattern (this provider is designed for Git-based distribution, not Registry publishing).

## Go Version Requirements

- **Go 1.24** (with toolchain go1.24.9)
- Uses generics and modern Go features
- Key dependencies: pulumi-go-provider v1.1.2, pulumi/sdk/v3 v3.169.0

## Provider Framework: pulumi-go-provider with infer

This provider uses the **infer framework** which:
- Automatically generates provider schema from Go types
- Derives component contracts from Go function signatures
- Reduces boilerplate compared to manual schema definition
- Uses reflection to map Go types to Pulumi types

The `infer.ComponentF()` wrapper adapts regular Go constructor functions to the provider interface.
