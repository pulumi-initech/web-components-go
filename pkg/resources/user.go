package resources

import (
	"context"

	"github.com/pulumi-initech/web-components-go/pkg/config"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type User struct{}
type UserArgs struct{}
type UserState struct {
	Name     string `pulumi:"name"`
	Password string `pulumi:"password"`
}

func (*User) Create(ctx context.Context, req infer.CreateRequest[UserArgs]) (infer.CreateResponse[UserState], error) {
	cfg := infer.GetConfig[config.Config](ctx)
	return infer.CreateResponse[UserState]{
		ID: req.Name,
		Output: UserState{
			Name:     cfg.User,
			Password: cfg.HashedPassword,
		}}, nil
}

var _ = (infer.CustomDiff[UserArgs, UserState])((*User)(nil))

func (*User) Diff(ctx context.Context, req infer.DiffRequest[UserArgs, UserState]) (infer.DiffResponse, error) {
	cfg := infer.GetConfig[config.Config](ctx)
	if cfg.User != req.State.Name {
		return infer.DiffResponse{
			HasChanges: true,
			DetailedDiff: map[string]p.PropertyDiff{
				"name": {Kind: p.UpdateReplace},
			},
		}, nil
	}
	return infer.DiffResponse{}, nil
}
