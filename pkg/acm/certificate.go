package acm

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DnsValidatedCertificateArgs defines the input arguments for the DnsValidatedCertificate component
type DnsValidatedCertificateArgs struct {
	DomainName pulumi.StringInput `pulumi:"domainName"`
	ZoneId     pulumi.StringInput `pulumi:"zoneId"`
}

// DnsValidatedCertificate is a component that creates an ACM certificate with DNS validation
type DnsValidatedCertificate struct {
	pulumi.ResourceState

	DnsValidatedCertificateArgs

	CertificateArn pulumi.StringOutput `pulumi:"certificateArn"`
}

// NewDnsValidatedCertificate creates a new DnsValidatedCertificate component
func NewDnsValidatedCertificate(ctx *pulumi.Context, name string, args DnsValidatedCertificateArgs, opts ...pulumi.ResourceOption) (*DnsValidatedCertificate, error) {
	comp := &DnsValidatedCertificate{}

	// Register the component resource
	err := ctx.RegisterComponentResource("web-components:acm:DnsValidatedCertificate", name, comp, opts...)
	if err != nil {
		return nil, err
	}

	// Create the ACM certificate
	cert, err := acm.NewCertificate(ctx, "default", &acm.CertificateArgs{
		DomainName:       args.DomainName,
		ValidationMethod: pulumi.String("DNS"),
	}, pulumi.Parent(comp))
	if err != nil {
		return nil, err
	}

	// Create the Route53 validation record
	certValidationRecord, err := route53.NewRecord(ctx, "domainName-valid", &route53.RecordArgs{
		Name: cert.DomainValidationOptions.Index(pulumi.Int(0)).ResourceRecordName().Elem(),
		Records: pulumi.StringArray{
			cert.DomainValidationOptions.Index(pulumi.Int(0)).ResourceRecordValue().Elem(),
		},
		Type:   cert.DomainValidationOptions.Index(pulumi.Int(0)).ResourceRecordType().Elem(),
		Ttl:    pulumi.Int(60),
		ZoneId: args.ZoneId,
	}, pulumi.Parent(comp))
	if err != nil {
		return nil, err
	}

	// Create the certificate validation resource
	certValidation, err := acm.NewCertificateValidation(ctx, "cert", &acm.CertificateValidationArgs{
		CertificateArn: cert.Arn,
		ValidationRecordFqdns: pulumi.StringArray{
			certValidationRecord.Fqdn,
		},
	}, pulumi.Parent(cert))
	if err != nil {
		return nil, err
	}

	// Set the output
	comp.CertificateArn = certValidation.CertificateArn

	// Register outputs
	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"certificateArn": comp.CertificateArn,
	}); err != nil {
		return nil, err
	}

	return comp, nil
}

// Annotate provides descriptions for the component and its properties
func (c *DnsValidatedCertificate) Annotate(a infer.Annotator) {
	a.Describe(&c, "Creates an ACM certificate with automatic DNS validation via Route53")
	a.Describe(&c.DomainName, "The domain name for which the certificate should be issued")
	a.Describe(&c.ZoneId, "The Route53 hosted zone ID where the validation record will be created")
	a.Describe(&c.CertificateArn, "The ARN of the validated certificate")
}
