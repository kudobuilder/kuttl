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

// TestCreateDeletePolicy_None verifies that DeleteNone registers no cleanup callbacks,
// leaving created objects alive after the test ends.
func TestCreateDeletePolicy_None(t *testing.T) {
	cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	s := newStepWithClient(t, cl, harness.DeleteNone)

	errs := s.Create(t, testNamespace)
	require.Empty(t, errs)

	// Since DeletePolicy is DeleteNone, no clean callback was
	// registered, and therefore the resource should not be deleted.
	pod := kubernetes.NewPod("test-pod", testNamespace)
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod))
}

// TestCreateDeletePolicy_All verifies that DeleteAll cleans up resources regardless of
// whether the step succeeded.
func TestCreateDeletePolicy_All(t *testing.T) {
	cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	s := newStepWithClient(t, cl, harness.DeleteAll)
	// s.succeeded remains false – DeleteAll must ignore it.

	errs := s.Create(t, testNamespace)
	require.Empty(t, errs)

	pod := kubernetes.NewPod("test-pod", testNamespace)
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod), "object should exist right after Create")

	// Simulate what the registered t.Cleanup would do via Clean().
	require.NoError(t, s.Clean(testNamespace))
	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod)),
		"DeleteAll: object should be gone after cleanup")
}

// TestCreateDeletePolicy_Success_OnSuccess verifies that DeleteSuccess deletes resources
// when s.succeeded is true at cleanup time.
func TestCreateDeletePolicy_Success_OnSuccess(t *testing.T) {
	var cl client.Client

	t.Run("inner", func(t *testing.T) {
		cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		s := newStepWithClient(t, cl, harness.DeleteSuccess)
		s.succeeded = true // simulate a passing step

		errs := s.Create(t, testNamespace)
		require.Empty(t, errs)

		pod := kubernetes.NewPod("test-pod", testNamespace)
		require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod))
		// t.Cleanup fires when the sub-test ends; succeeded==true → deletion runs.
	})

	// After the sub-test cleanup the object must be gone.
	pod := kubernetes.NewPod("test-pod", testNamespace)
	assert.True(t, k8serrors.IsNotFound(cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod)),
		"DeleteSuccess+succeeded=true: object should be deleted after test cleanup")
}

// TestCreateDeletePolicy_Success_OnFailure verifies that DeleteSuccess preserves resources
// when s.succeeded is false at cleanup time.
func TestCreateDeletePolicy_Success_OnFailure(t *testing.T) {
	var cl client.Client

	t.Run("inner", func(t *testing.T) {
		cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		s := newStepWithClient(t, cl, harness.DeleteSuccess)
		// s.succeeded remains false (default) – cleanup closure must skip deletion.

		errs := s.Create(t, testNamespace)
		require.Empty(t, errs)

		pod := kubernetes.NewPod("test-pod", testNamespace)
		require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod))
		// t.Cleanup fires when the sub-test ends; succeeded==false → deletion skipped.
	})

	// After the sub-test the object must still be present.
	pod := kubernetes.NewPod("test-pod", testNamespace)
	require.NoError(t, cl.Get(t.Context(), kubernetes.ObjectKey(pod), pod),
		"DeleteSuccess+succeeded=false: object should survive after test cleanup")
}
