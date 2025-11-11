package config

import (
	"context"
	"fmt"
	"hash/adler32"
	"hash/crc32"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type HashKind string

// var _ = (infer.Enum[HashKind])((*HashKind)(nil))

const (
	HashAdler HashKind = "Adler32"
	HashCRC   HashKind = "CRC32"
)

type Config struct {
	User     string   `pulumi:"user"`
	Password string   `pulumi:"password,optional" provider:"secret"`
	HashKind HashKind `pulumi:"hash"`

	HashedPassword string
}

var _ = (infer.Annotated)((*Config)(nil))

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.User, "The username. Its important but not secret.")
	a.Describe(&c.Password, "The password. It is very secret.")
	a.Describe(&c.HashKind, `The (entirely uncryptographic) hash function used to encode the "password".`)
	a.SetDefault(&c.Password, "", "FOO")
	a.SetDefault(&c.HashKind, HashAdler)
}

var _ = (infer.CustomConfigure)((*Config)(nil))

func (c *Config) Configure(ctx context.Context) error {
	msg := fmt.Sprintf("credentials provider setup with user: %q", c.User)
	if c.Password != "" {
		msg += fmt.Sprintf(" and a very secret password (its %q)", c.Password)
	}
	switch c.HashKind {
	case HashAdler:
		c.HashedPassword = fmt.Sprintf("%d", adler32.Checksum([]byte(c.Password)))
	case HashCRC:
		c.HashedPassword = fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(c.Password)))
	}
	p.GetLogger(ctx).Info(msg)
	return nil
}
