package step

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kudobuilder/kuttl/internal/kubernetes"
	k8sfake "github.com/kudobuilder/kuttl/internal/kubernetes/fake"
	testutils "github.com/kudobuilder/kuttl/internal/utils"
	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

// newStepWithClient builds a minimal Step wired to the supplied fake client.
func newStepWithClient(t *testing.T, cl client.Client, policy harness.DeletePolicy) *Step {
	t.Helper()
	return &Step{
		DeletePolicy: policy,
		Logger:       testutils.NewTestLogger(t, ""),
		Client:       func(bool) (client.Client, error) { return cl, nil },
		DiscoveryClient: func() (discovery.DiscoveryInterface, error) {
			return k8sfake.DiscoveryClient(), nil
		},
		Apply: []client.Object{
			kubernetes.NewPod("test-pod", ""),
		},
	}
}

// TestCreateDeletePolicy verifies that Step.Create registers a cleanup callback that
// honours the configured DeletePolicy (and, for DeleteSuccess, the step's succeeded state)
// once the sub-test in which Create ran has ended.
func TestCreateDeletePolicy(t *testing.T) {
	tests := []struct {
		name          string
		policy        harness.DeletePolicy
		succeeded     bool
		shouldSurvive bool
	}{
		{"None never deletes on failure", harness.DeleteNone, false, true},
		{"None never deletes on success", harness.DeleteNone, true, true},
		{"All ignores succeeded on failure", harness.DeleteAll, false, false},
		{"All ignores succeeded on success", harness.DeleteAll, true, false},
		{"Success deletes on success", harness.DeleteSuccess, true, false},
		{"Success keeps on failure", harness.DeleteSuccess, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

			t.Run("inner", func(t *testing.T) {
				s := newStepWithClient(t, cl, tt.policy)
				s.succeeded = tt.succeeded

				errs := s.Create(t, testNamespace)
				require.Empty(t, errs)

				// The object exists once Create succeeds, before cleanup runs.
				pod := kubernetes.NewPod("test-pod", testNamespace)
				require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod))
				// t.Cleanup fires when this sub-test ends, applying the delete policy.
			})

			// After the sub-test cleanup, check whether the object survived.
			pod := kubernetes.NewPod("test-pod", testNamespace)
			err := cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod)
			if tt.shouldSurvive {
				require.NoError(t, err, "object should survive after test cleanup")
			} else {
				assert.True(t, k8serrors.IsNotFound(err), "object should be gone after test cleanup")
			}
		})
	}
}
