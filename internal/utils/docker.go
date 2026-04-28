package utils //nolint:revive,nolintlint // apparently nolintlint is confused

import (
	"context"

	"github.com/moby/moby/client"
)

// DockerClient is a wrapper interface for the Docker library to support unit testing.
type DockerClient interface {
	VolumeCreate(context.Context, client.VolumeCreateOptions) (client.VolumeCreateResult, error)
	ImageSave(context.Context, []string, ...client.ImageSaveOption) (client.ImageSaveResult, error)
}
