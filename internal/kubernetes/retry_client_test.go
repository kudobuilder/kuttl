package kubernetes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRetry(t *testing.T) {
	index := 0

	assert.Nil(t, retry(context.TODO(), func(context.Context) error {
		index++
		if index == 1 {
			return errors.New("ignore this error")
		}
		return nil
	}, func(error) bool { return false }, func(err error) bool {
		return err.Error() == "ignore this error"
	}))

	assert.Equal(t, 2, index)
}

func TestRetryWithUnexpectedError(t *testing.T) {
	index := 0

	assert.Equal(t, errors.New("bad error"), retry(context.TODO(), func(context.Context) error {
		index++
		if index == 1 {
			return errors.New("bad error")
		}
		return nil
	}, func(error) bool { return false }, func(err error) bool {
		return err.Error() == "ignore this error"
	}))
	assert.Equal(t, 1, index)
}

func TestRetryWithNil(t *testing.T) {
	assert.Equal(t, nil, retry(context.TODO(), nil, isJSONSyntaxError))
}

func TestRetryWithNilFromFn(t *testing.T) {
	assert.Equal(t, nil, retry(context.TODO(), func(context.Context) error {
		return nil
	}, isJSONSyntaxError))
}

func TestRetryWithNilInFn(t *testing.T) {
	c := RetryClient{}
	var list client.ObjectList
	assert.Error(t, retry(context.TODO(), func(ctx context.Context) error {
		return c.Client.List(ctx, list)
	}, isJSONSyntaxError))
}

func TestRetryWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	assert.Equal(t, errors.New("error"), retry(ctx, func(context.Context) error {
		return errors.New("error")
	}, func(error) bool { return true }))
}
