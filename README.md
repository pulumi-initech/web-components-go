# Web Components Go Provider

A Pulumi component library written in Go for creating AWS web hosting environments with auto-scaling, load balancing, and SSL certificate management.

This project is a Go port of the Java-based `aws-webautoscaling-java` components, built using the [pulumi-go-provider](https://github.com/pulumi/pulumi-go-provider) framework.

## Components

### DnsValidatedCertificate

Creates an AWS Certificate Manager (ACM) certificate with automatic DNS validation via Route53.

**Type:** `web-components:index:DnsValidatedCertificate`

**Features:**
- Automatic ACM certificate creation
- DNS validation via Route53
- Waits for certificate validation to complete

**Inputs:**
- `domainName` (string, required): The domain name for which the certificate should be issued
- `zoneId` (string, required): The Route53 hosted zone ID where the validation record will be created

**Outputs:**
- `certificateArn` (string): The ARN of the validated certificate

### WebEnvironment

Creates a complete web hosting environment with auto-scaling, load balancing, and optional DNS configuration.

**Type:** `web-components:index:WebEnvironment`

**Features:**
- Auto Scaling Group with Launch Template
- Application Load Balancer with HTTPS termination
- Security groups for ALB and EC2 instances
- SSH key pair generation
- Nginx web server installation via user data
- CloudWatch alarms for request-based auto-scaling
- Optional Route53 DNS alias record

**Inputs:**
- `vpcId` (string, required): The VPC ID where resources will be created
- `vpcCidr` (string, required): The CIDR block of the VPC for security group rules
- `imageId` (string, required): The AMI ID to use for EC2 instances
- `instanceType` (string, required): The EC2 instance type
- `publicSubnetIds` (array of strings, required): List of public subnet IDs for the load balancer and instances
- `certificateArn` (string, required): The ARN of the ACM certificate for HTTPS
- `instanceCount` (integer, optional): The number of instances to run
- `privateSubnetIds` (array of strings, optional): List of private subnet IDs
- `zoneId` (string, optional): The Route53 hosted zone ID for DNS alias
- `subdomain` (string, optional): The subdomain for the DNS alias record

## Prerequisites

- Go 1.23 or later
- Pulumi CLI
- AWS credentials configured

## Building

To build the provider:

```bash
make build
```

To install the provider locally for testing:

```bash
make install
```

This will install the provider to `~/.pulumi/plugins/resource-web-components-v0.1.0/`.

## Development

### Project Structure

```
.
├── provider/
│   ├── cmd/
│   │   └── pulumi-resource-web-components/
│   │       └── main.go              # Provider entry point
│   ├── pkg/
│   │   ├── acm/
│   │   │   └── certificate.go       # DnsValidatedCertificate component
│   │   └── web/
│   │       └── environment.go       # WebEnvironment component
│   └── schema.json                  # Provider schema
├── go.mod                           # Go module definition
├── Makefile                         # Build automation
├── PulumiPlugin.yaml               # Pulumi plugin metadata
└── README.md                        # This file
```

### Available Make Targets

- `make build` - Build the provider binary
- `make install` - Build and install the provider locally
- `make clean` - Remove build artifacts and installed provider
- `make test` - Run tests
- `make deps` - Download and tidy dependencies
- `make fmt` - Format Go code
- `make lint` - Run linter (requires golangci-lint)

## Usage Examples

### Using DnsValidatedCertificate

```go
package main

import (
    "github.com/pulumi-initech/web-components-go/provider/pkg/acm"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cert, err := acm.NewDnsValidatedCertificate(ctx, "my-cert", acm.DnsValidatedCertificateArgs{
            DomainName: pulumi.String("example.com"),
            ZoneId:     pulumi.String("Z1234567890ABC"),
        })
        if err != nil {
            return err
        }

        ctx.Export("certificateArn", cert.CertificateArn)
        return nil
    })
}
```

### Using WebEnvironment

```go
package main

import (
    "github.com/pulumi-initech/web-components-go/provider/pkg/web"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        env, err := web.NewWebEnvironment(ctx, "my-web-env", web.WebEnvironmentArgs{
            VpcId:           pulumi.String("vpc-12345"),
            VpcCidr:         pulumi.String("10.0.0.0/16"),
            ImageId:         pulumi.String("ami-12345678"),
            InstanceType:    pulumi.String("t3.micro"),
            PublicSubnetIds: pulumi.StringArray{
                pulumi.String("subnet-1"),
                pulumi.String("subnet-2"),
            },
            CertificateArn: pulumi.String("arn:aws:acm:us-east-1:123456789012:certificate/..."),
            ZoneId:         pulumi.String("Z1234567890ABC"),
            Subdomain:      pulumi.String("www.example.com"),
        })
        if err != nil {
            return err
        }

        return nil
    })
}
```

## Comparison with Java Version

This Go implementation provides the same functionality as the Java version with the following differences:

1. **Language**: Written in Go instead of Java
2. **Framework**: Uses `pulumi-go-provider` instead of Java's `ComponentProviderHost`
3. **Structure**: Follows Go package conventions and idiomatic Go patterns
4. **Build System**: Uses Make instead of Gradle
5. **Performance**: Generally faster startup and lower memory footprint

## Contributing

Contributions are welcome! Please ensure that:

1. Code is properly formatted (`make fmt`)
2. All tests pass (`make test`)
3. Code follows Go best practices

## License

This project follows the same license as the original Java implementation.

## References

- [Pulumi Go Provider SDK](https://github.com/pulumi/pulumi-go-provider)
- [Original Java Implementation](https://github.com/pulumi-initech/aws-webautoscaling-java)
- [Pulumi Documentation](https://www.pulumi.com/docs/)
