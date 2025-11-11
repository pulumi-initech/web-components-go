# Web Components Go Provider

A Pulumi component library written in Go for creating AWS web hosting environments with auto-scaling, load balancing, and SSL certificate management.

Built using the [pulumi-go-provider](https://github.com/pulumi/pulumi-go-provider) framework.

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

**Outputs:**
- `loadBalancerUrl` (string): The HTTPS URL of the Application Load Balancer

## Prerequisites

- Go 1.24 or later
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
├── main.go                          # Provider entry point
├── pkg/
│   ├── acm/
│   │   └── certificate.go           # DnsValidatedCertificate component
│   └── web/
│       └── environment.go           # WebEnvironment component
├── go.mod                           # Go module definition
├── Makefile                         # Build automation
├── PulumiPlugin.yaml               # Pulumi plugin metadata
├── schema.json                      # Provider schema
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

## Usage Example

This example shows how to create a complete web hosting environment with SSL certificate:

```go
package main

import (
    "github.com/pulumi-initech/web-components-go/pkg/acm"
    "github.com/pulumi-initech/web-components-go/pkg/web"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        // Create an SSL certificate with DNS validation
        cert, err := acm.NewDnsValidatedCertificate(ctx, "ssl-cert", acm.DnsValidatedCertificateArgs{
            DomainName: pulumi.String("www.example.com"),
            ZoneId:     pulumi.String("Z1234567890ABC"),
        })
        if err != nil {
            return err
        }

        // Create the web hosting environment using the certificate
        env, err := web.NewWebEnvironment(ctx, "web-env", web.WebEnvironmentArgs{
            VpcId:           pulumi.String("vpc-12345678"),
            VpcCidr:         pulumi.String("10.0.0.0/16"),
            ImageId:         pulumi.String("ami-12345678"),
            InstanceType:    pulumi.String("t3.micro"),
            PublicSubnetIds: pulumi.StringArray{
                pulumi.String("subnet-abc123"),
                pulumi.String("subnet-def456"),
            },
            CertificateArn:  cert.CertificateArn,
            ZoneId:          cert.ZoneId,
            Subdomain:       cert.DomainName,
        })
        if err != nil {
            return err
        }

        // Export outputs
        ctx.Export("certificateArn", cert.CertificateArn)
        ctx.Export("loadBalancerUrl", env.LoadBalancerUrl)

        return nil
    })
}
```

This example demonstrates the typical workflow:

1. Create a DNS-validated SSL certificate for your domain
2. Create a web environment that uses the certificate for HTTPS
3. The web environment automatically sets up auto-scaling, load balancing, and DNS records

## Contributing

Contributions are welcome! Please ensure that:

1. Code is properly formatted (`make fmt`)
2. All tests pass (`make test`)
3. Code follows Go best practices

## References

- [Pulumi Go Provider SDK](https://github.com/pulumi/pulumi-go-provider)
- [Pulumi Documentation](https://www.pulumi.com/docs/)
