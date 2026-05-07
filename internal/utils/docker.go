package utils //nolint:revive,nolintlint // apparently nolintlint is confused

import (
	"context"

	"github.com/moby/moby/client"
)

// DockerClient is a wrapper interface for the Docker library to support unit testing.
type DockerClient interface {
	VolumeCreate(ctx context.Context, options client.VolumeCreateOptions) (client.VolumeCreateResult, error)
	ImageSave(ctx context.Context, imageIDs []string, opts ...client.ImageSaveOption) (client.ImageSaveResult, error)
}
