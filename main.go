package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi-initech/web-components-go/acm"
	"github.com/pulumi-initech/web-components-go/web"
	"github.com/pulumi/pulumi-go-provider/infer"
)

const (
	providerName    = "web-components"
	providerVersion = "0.1.0"
)

func main() {
	// Build the provider with both component resources
	provider, err := infer.NewProviderBuilder().
		WithComponents(
			infer.ComponentF(acm.NewDnsValidatedCertificate),
			infer.ComponentF(web.NewWebEnvironment),
		).
		Build()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building provider: %v\n", err)
		os.Exit(1)
	}

	// Run the provider
	provider.Run(context.Background(), providerName, providerVersion)
}
