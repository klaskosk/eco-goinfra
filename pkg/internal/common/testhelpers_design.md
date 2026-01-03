# Design Document: Common Test Helpers for Pull Functions

## Problem Statement

The common package provides generic `PullClusterScopedBuilder` and `PullNamespacedBuilder` functions that are well-tested. However, resource-specific packages (like `bmh`) wrap these functions in their own Pull functions (e.g., `PullHFS`, `PullHFC`).

These wrappers are thin, but they can still have bugs:

- Arguments passed in wrong order (e.g., `name` and `nsname` swapped)
- Wrong scheme attacher used
- Incorrect GVK returned from builder

Black-box testing of these wrappers provides assurance that the integration is correct.

## Goals

1. Create reusable test helpers that can validate any Pull function wrapper
2. Minimize boilerplate in individual package tests
3. Catch common integration mistakes (argument swapping, wrong scheme, etc.)
4. Leverage Go generics to work with any builder type

## Design Overview

### Type Signatures

The Pull functions have different signatures depending on scope:

```go
import runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

// Namespaced Pull function signature (e.g., PullHFS, PullHFC).
// C is generic over the API client type (e.g., *clients.Settings or runtimeclient.Client).
type NamespacedPullFunc[C runtimeclient.Client, SB any] func(apiClient C, name, nsname string) (SB, error)

// Cluster-scoped Pull function signature.
type ClusterScopedPullFunc[C runtimeclient.Client, SB any] func(apiClient C, name string) (SB, error)
```

### Required Test Configuration

To test a Pull function, we need:

1. **A way to create a test object** - A function that creates a dummy K8s resource for seeding the fake client
2. **The scheme attacher** - To register the resource types with the fake client
3. **Expected GVK** - To verify the builder is configured correctly
4. **Known name/namespace values** - To verify arguments are passed correctly

### Proposed Interface

```go
import (
    "github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
    "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
    "k8s.io/apimachinery/pkg/runtime/schema"
    runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectPointer mirrors the shape of the common package's internal object pointer constraint.
// Re-declaring it here lets test helpers live in a separate package without exporting common internals.
type ObjectPointer[O any] interface {
    *O
    runtimeclient.Object
}

// BuilderPointer constrains SB to be a pointer-to-builder that implements common.Builder for (O, SO).
type BuilderPointer[B, O any, SO ObjectPointer[O]] interface {
    *B
    common.Builder[O, SO]
}

// NamespacedPullTestConfig provides the configuration needed to test a namespaced Pull function wrapper.
type NamespacedPullTestConfig[
    C runtimeclient.Client,
    O, B any,
    SO ObjectPointer[O],
    SB BuilderPointer[B, O, SO],
] struct {
    // PullFunc is the Pull wrapper function to test.
    PullFunc NamespacedPullFunc[C, SB]

    // ClientFactory builds a test API client of type C.
    ClientFactory func(clients.TestClientParams) C

    // SchemeAttacher registers the resource type with the client scheme (required for seeding existing objects).
    SchemeAttacher clients.SchemeAttacher

    // BuildTestObject creates a test object with the given name and namespace for seeding the fake client.
    BuildTestObject func(name, nsname string) SO

    // ExpectedGVK is the expected GVK for the builder.
    ExpectedGVK schema.GroupVersionKind
}

// ClusterScopedPullTestConfig provides the configuration needed to test a cluster-scoped Pull function wrapper.
type ClusterScopedPullTestConfig[
    C runtimeclient.Client,
    O, B any,
    SO ObjectPointer[O],
    SB BuilderPointer[B, O, SO],
] struct {
    PullFunc ClusterScopedPullFunc[C, SB]

    ClientFactory func(clients.TestClientParams) C

    SchemeAttacher clients.SchemeAttacher

    BuildTestObject func(name string) SO

    ExpectedGVK schema.GroupVersionKind
}
```

### Test Cases Covered

The test helper should run these test cases:

| Test Case | Purpose | Expected Result |
|-----------|---------|-----------------|
| Valid pull of existing resource | Happy path | Builder returned, definition/object populated |
| Nil client | Validate nil handling | Error: API client nil |
| Empty name | Validate name required | Error: Name empty |
| Empty namespace (namespaced only) | Validate namespace required | Error: Namespace empty |
| Resource does not exist (**no pre-attached scheme**) | Verify wrapper attaches correct scheme and passes correct args | Error: Not found |

In addition, the helper should perform a **scheme attacher/GVK preflight**: run `SchemeAttacher` on a fresh `runtime.Scheme`, then assert that `scheme.ObjectKinds(testObject)` includes `ExpectedGVK`.

### Implementation Sketch

```go
package testhelpers

import (
    "testing"

    "github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
    "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
    commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
    "github.com/stretchr/testify/assert"
    k8serrors "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
    runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
    TestResourceName      = "test-resource-name"
    TestResourceNamespace = "test-resource-namespace"
)

// ObjectPointer mirrors the shape of the common package's internal object pointer constraint.
// Re-declaring it here lets test helpers live in a separate package without exporting common internals.
type ObjectPointer[O any] interface {
    *O
    runtimeclient.Object
}

// BuilderPointer constrains SB to be a pointer-to-builder that implements common.Builder for (O, SO).
type BuilderPointer[B, O any, SO ObjectPointer[O]] interface {
    *B
    common.Builder[O, SO]
}

// NamespacedPullFunc represents a function that pulls a namespaced resource.
// C is generic over the API client type (e.g., *clients.Settings or runtimeclient.Client).
type NamespacedPullFunc[C runtimeclient.Client, SB any] func(apiClient C, name, nsname string) (SB, error)

// ClusterScopedPullFunc represents a function that pulls a cluster-scoped resource.
// C is generic over the API client type (e.g., *clients.Settings or runtimeclient.Client).
type ClusterScopedPullFunc[C runtimeclient.Client, SB any] func(apiClient C, name string) (SB, error)

// NamespacedPullTestConfig configures the test for a namespaced Pull function.
type NamespacedPullTestConfig[C runtimeclient.Client, O, B any, SO ObjectPointer[O], SB BuilderPointer[B, O, SO]] struct {
    PullFunc        NamespacedPullFunc[C, SB]
    ClientFactory   func(clients.TestClientParams) C
    SchemeAttacher  clients.SchemeAttacher
    BuildTestObject func(name, nsname string) SO
    ExpectedGVK     schema.GroupVersionKind
}

// ClusterScopedPullTestConfig configures the test for a cluster-scoped Pull function wrapper.
type ClusterScopedPullTestConfig[C runtimeclient.Client, O, B any, SO ObjectPointer[O], SB BuilderPointer[B, O, SO]] struct {
    PullFunc        ClusterScopedPullFunc[C, SB]
    ClientFactory   func(clients.TestClientParams) C
    SchemeAttacher  clients.SchemeAttacher
    BuildTestObject func(name string) SO
    ExpectedGVK     schema.GroupVersionKind
}

// NewClusterScopedPullTestConfig is a type-inference-friendly constructor that avoids explicit type parameters at call sites.
func NewClusterScopedPullTestConfig[C runtimeclient.Client, O, B any, SO ObjectPointer[O], SB BuilderPointer[B, O, SO]](
    pullFunc ClusterScopedPullFunc[C, SB],
    clientFactory func(clients.TestClientParams) C,
    schemeAttacher clients.SchemeAttacher,
    buildTestObject func(name string) SO,
    expectedGVK schema.GroupVersionKind,
) ClusterScopedPullTestConfig[C, O, B, SO, SB] {
    return ClusterScopedPullTestConfig[C, O, B, SO, SB]{
        PullFunc:        pullFunc,
        ClientFactory:   clientFactory,
        SchemeAttacher:  schemeAttacher,
        BuildTestObject: buildTestObject,
        ExpectedGVK:     expectedGVK,
    }
}

// NewNamespacedPullTestConfig is a type-inference-friendly constructor that avoids explicit type parameters at call sites.
func NewNamespacedPullTestConfig[C runtimeclient.Client, O, B any, SO ObjectPointer[O], SB BuilderPointer[B, O, SO]](
    pullFunc NamespacedPullFunc[C, SB],
    clientFactory func(clients.TestClientParams) C,
    schemeAttacher clients.SchemeAttacher,
    buildTestObject func(name, nsname string) SO,
    expectedGVK schema.GroupVersionKind,
) NamespacedPullTestConfig[C, O, B, SO, SB] {
    return NamespacedPullTestConfig[C, O, B, SO, SB]{
        PullFunc:        pullFunc,
        ClientFactory:   clientFactory,
        SchemeAttacher:  schemeAttacher,
        BuildTestObject: buildTestObject,
        ExpectedGVK:     expectedGVK,
    }
}

// RunNamespacedPullTests runs a comprehensive test suite for a namespaced Pull function wrapper.
func RunNamespacedPullTests[C runtimeclient.Client, O, B any, SO ObjectPointer[O], SB BuilderPointer[B, O, SO]](
    t *testing.T,
    config NamespacedPullTestConfig[C, O, B, SO, SB],
) {
    t.Helper()

    // Preflight: validate the SchemeAttacher/ExpectedGVK pair using runtime.Scheme.ObjectKinds.
    assertSchemeAttacherRegistersGVK(t, config.SchemeAttacher, config.BuildTestObject, config.ExpectedGVK)

    testCases := []struct {
        name          string
        clientNil     bool
        builderName   string
        builderNsName string
        objectExists  bool
        assertError   func(error) bool
    }{
        {
            name:          "valid pull existing resource",
            clientNil:     false,
            builderName:   TestResourceName,
            builderNsName: TestResourceNamespace,
            objectExists:  true,
            assertError:   isErrorNil,
        },
        {
            name:          "nil client returns error",
            clientNil:     true,
            builderName:   TestResourceName,
            builderNsName: TestResourceNamespace,
            objectExists:  false,
            assertError:   commonerrors.IsAPIClientNil,
        },
        {
            name:          "empty name returns error",
            clientNil:     false,
            builderName:   "",
            builderNsName: TestResourceNamespace,
            objectExists:  false,
            assertError:   commonerrors.IsBuilderNameEmpty,
        },
        {
            name:          "empty namespace returns error",
            clientNil:     false,
            builderName:   TestResourceName,
            builderNsName: "",
            objectExists:  false,
            assertError:   commonerrors.IsBuilderNamespaceEmpty,
        },
        {
            name:          "non-existent resource returns not found (no pre-attached scheme)",
            clientNil:     false,
            builderName:   "non-existent-resource",
            builderNsName: "non-existent-namespace",
            objectExists:  false,
            assertError:   k8serrors.IsNotFound,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.name, func(t *testing.T) {
            var (
                client  C
                objects []runtime.Object
            )

            if !testCase.clientNil {
                if testCase.objectExists {
                    // Use the exact name/namespace that will be queried
                    objects = append(objects, config.BuildTestObject(testCase.builderName, testCase.builderNsName))
                }

                // Only pre-attach the scheme when we must seed typed objects into the fake client.
                // For "not found" cases, leaving SchemeAttachers empty makes the test catch wrappers that use
                // the wrong scheme attacher (it would fail with "no kind is registered..." instead of NotFound).
                schemeAttachers := []clients.SchemeAttacher(nil)
                if testCase.objectExists {
                    schemeAttachers = []clients.SchemeAttacher{config.SchemeAttacher}
                }

                assert.NotNil(t, config.ClientFactory, "ClientFactory must be provided for non-nil client test cases")
                if config.ClientFactory == nil {
                    return
                }

                client = config.ClientFactory(clients.TestClientParams{
                    K8sMockObjects:  objects,
                    SchemeAttachers: schemeAttachers,
                })
            }

            builder, err := config.PullFunc(client, testCase.builderName, testCase.builderNsName)

            assert.Truef(t, testCase.assertError(err), "got error %v", err)

            if err == nil {
                assert.NotNil(t, builder)
                assert.NotNil(t, builder.GetDefinition())

                // Verify name/namespace are correctly assigned (catches argument swapping bugs).
                assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName(),
                    "definition name mismatch - check argument order in Pull function")
                assert.Equal(t, testCase.builderNsName, builder.GetDefinition().GetNamespace(),
                    "definition namespace mismatch - check argument order in Pull function")

                // Verify object is populated as well.
                assert.NotNil(t, builder.GetObject(), "object should be populated after pull")
                assert.Equal(t, testCase.builderName, builder.GetObject().GetName())
                assert.Equal(t, testCase.builderNsName, builder.GetObject().GetNamespace())

                // Verify GVK is correctly set.
                assert.Equal(t, config.ExpectedGVK, builder.GetGVK(),
                    "builder GVK mismatch - check GetGVK implementation")
            } else {
                assert.Nil(t, builder)
            }
        })
    }
}

// RunClusterScopedPullTests runs a comprehensive test suite for a cluster-scoped Pull function wrapper.
func RunClusterScopedPullTests[C runtimeclient.Client, O, B any, SO ObjectPointer[O], SB BuilderPointer[B, O, SO]](
    t *testing.T,
    config ClusterScopedPullTestConfig[C, O, B, SO, SB],
) {
    t.Helper()

    // Preflight: validate the SchemeAttacher/ExpectedGVK pair using runtime.Scheme.ObjectKinds.
    assertClusterScopedSchemeAttacherRegistersGVK(t, config.SchemeAttacher, config.BuildTestObject, config.ExpectedGVK)

    testCases := []struct {
        name         string
        clientNil    bool
        builderName  string
        objectExists bool
        assertError  func(error) bool
    }{
        {
            name:         "valid pull existing resource",
            clientNil:    false,
            builderName:  TestResourceName,
            objectExists: true,
            assertError:  isErrorNil,
        },
        {
            name:         "nil client returns error",
            clientNil:    true,
            builderName:  TestResourceName,
            objectExists: false,
            assertError:  commonerrors.IsAPIClientNil,
        },
        {
            name:         "empty name returns error",
            clientNil:    false,
            builderName:  "",
            objectExists: false,
            assertError:  commonerrors.IsBuilderNameEmpty,
        },
        {
            name:         "non-existent resource returns not found (no pre-attached scheme)",
            clientNil:    false,
            builderName:  "non-existent-resource",
            objectExists: false,
            assertError:  k8serrors.IsNotFound,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.name, func(t *testing.T) {
            var (
                client  C
                objects []runtime.Object
            )

            if !testCase.clientNil {
                if testCase.objectExists {
                    objects = append(objects, config.BuildTestObject(testCase.builderName))
                }

                // Only pre-attach the scheme when we must seed typed objects into the fake client.
                schemeAttachers := []clients.SchemeAttacher(nil)
                if testCase.objectExists {
                    schemeAttachers = []clients.SchemeAttacher{config.SchemeAttacher}
                }

                assert.NotNil(t, config.ClientFactory, "ClientFactory must be provided for non-nil client test cases")
                if config.ClientFactory == nil {
                    return
                }

                client = config.ClientFactory(clients.TestClientParams{
                    K8sMockObjects:  objects,
                    SchemeAttachers: schemeAttachers,
                })
            }

            builder, err := config.PullFunc(client, testCase.builderName)
            assert.Truef(t, testCase.assertError(err), "got error %v", err)

            if err == nil {
                assert.NotNil(t, builder)
                assert.NotNil(t, builder.GetDefinition())

                assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName(),
                    "definition name mismatch - check argument order in Pull function")

                assert.NotNil(t, builder.GetObject(), "object should be populated after pull")
                assert.Equal(t, testCase.builderName, builder.GetObject().GetName())

                assert.Equal(t, config.ExpectedGVK, builder.GetGVK(),
                    "builder GVK mismatch - check GetGVK implementation")
            } else {
                assert.Nil(t, builder)
            }
        })
    }
}

func assertSchemeAttacherRegistersGVK[O any, SO ObjectPointer[O]](
    t *testing.T,
    schemeAttacher clients.SchemeAttacher,
    buildTestObject func(name, nsname string) SO,
    expectedGVK schema.GroupVersionKind,
) {
    t.Helper()

    scheme := runtime.NewScheme()
    err := schemeAttacher(scheme)
    assert.NoError(t, err, "schemeAttacher failed when attaching to a fresh scheme")
    if err != nil {
        return
    }

    obj := buildTestObject(TestResourceName, TestResourceNamespace)
    kinds, _, err := scheme.ObjectKinds(obj)
    assert.NoError(t, err, "scheme.ObjectKinds failed for test object; scheme attacher may be wrong")
    assert.Contains(t, kinds, expectedGVK, "scheme attacher did not register the expected GVK")
}

func assertClusterScopedSchemeAttacherRegistersGVK[O any, SO ObjectPointer[O]](
    t *testing.T,
    schemeAttacher clients.SchemeAttacher,
    buildTestObject func(name string) SO,
    expectedGVK schema.GroupVersionKind,
) {
    t.Helper()

    scheme := runtime.NewScheme()
    err := schemeAttacher(scheme)
    assert.NoError(t, err, "schemeAttacher failed when attaching to a fresh scheme")
    if err != nil {
        return
    }

    obj := buildTestObject(TestResourceName)
    kinds, _, err := scheme.ObjectKinds(obj)
    assert.NoError(t, err, "scheme.ObjectKinds failed for test object; scheme attacher may be wrong")
    assert.Contains(t, kinds, expectedGVK, "scheme attacher did not register the expected GVK")
}

func isErrorNil(err error) bool {
    return err == nil
}
```

### Usage Example

```go
// pkg/bmh/hfs_test.go
package bmh

import (
    "testing"

    bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
    "github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
    "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelpers"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPullHFS(t *testing.T) {
    t.Parallel()

    cfg := testhelpers.NewNamespacedPullTestConfig(
        PullHFS,
        clients.GetTestClients,
        bmhv1alpha1.AddToScheme,
        func(name, nsname string) *bmhv1alpha1.HostFirmwareSettings {
            return &bmhv1alpha1.HostFirmwareSettings{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      name,
                    Namespace: nsname,
                },
            }
        },
        bmhv1alpha1.GroupVersion.WithKind("HostFirmwareSettings"),
    )

    testhelpers.RunNamespacedPullTests(t, cfg)
}

func TestPullHFC(t *testing.T) {
    t.Parallel()

    cfg := testhelpers.NewNamespacedPullTestConfig(
        PullHFC,
        clients.GetTestClients,
        bmhv1alpha1.AddToScheme,
        func(name, nsname string) *bmhv1alpha1.HostFirmwareComponents {
            return &bmhv1alpha1.HostFirmwareComponents{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      name,
                    Namespace: nsname,
                },
            }
        },
        bmhv1alpha1.GroupVersion.WithKind("HostFirmwareComponents"),
    )

    testhelpers.RunNamespacedPullTests(t, cfg)
}
```

## Reducing Boilerplate: Type Inference-Friendly API

The explicit generic parameters are intentionally pushed into `NewNamespacedPullTestConfig` so callers don't need to spell them out. This keeps the implementation fully generic while keeping call sites compact.

## Key Design Decisions

### 1. Why not just trust the underlying common.PullNamespacedBuilder?

While `PullNamespacedBuilder` is tested, the wrapper functions can still have bugs:

- Swapped arguments
- Wrong scheme attacher
- Incorrect context usage
- Future refactoring mistakes

### 2. Why require BuildTestObject instead of using reflection?

Explicit object construction:

- Is type-safe at compile time
- Allows tests to set any required fields for the resource
- Avoids reflection complexity
- Makes it clear what object is being tested

### 3. Why go “all-in” on generics and reuse common.Builder?

- **No adapter wrappers needed**: wrapper pull functions like `func(*clients.Settings, string, string) (*HFSBuilder, error)` can be passed directly.
- **Leverages the repo’s canonical builder contract**: `common.Builder[O, SO]` is already the shared surface for `GetDefinition`, `GetObject`, and `GetGVK`.
- **Avoids interface return-type pitfalls**: builder methods in this repo return the concrete pointer type `SO` (e.g., `*corev1.ConfigMap`), so an interface that requires `GetDefinition() runtimeclient.Object` cannot be satisfied (Go has no covariant return types).

### 4. Why separate configs for namespaced vs cluster-scoped?

- Different Pull function signatures (with/without namespace)
- Different test cases (namespace-related tests only for namespaced)
- Clearer intent in test code

## Files to Create

1. `pkg/internal/common/testhelpers/pull.go` - Main test helper implementations
2. `pkg/internal/common/testhelpers/doc.go` - Package documentation

## Testing the Test Helpers

The helpers themselves can be tested with intentionally buggy pull functions. You can avoid a mock `testing.T` by using `t.Run`'s boolean return value:

```go
func TestRunNamespacedPullTests_CatchesSwappedArguments(t *testing.T) {
    ok := t.Run("RunNamespacedPullTests should fail", func(t *testing.T) {
        cfg := NewNamespacedPullTestConfig(
            buggyPullFunc,        // swaps name/ns when delegating
            clients.GetTestClients,
            corev1.AddToScheme,
            func(name, ns string) *corev1.ConfigMap { /* ... */ },
            corev1.SchemeGroupVersion.WithKind("ConfigMap"),
        )

        RunNamespacedPullTests(t, cfg)
    })

    if ok {
        t.Fatalf("expected helper to fail the subtest, but it passed")
    }
}
```

## Technical Considerations

### Unexported type constraints in common

The `common` package uses unexported constraints (`objectPointer`, `builderPointer`). To keep test helpers in a separate package *without* exporting those internals, the helper package can re-declare equivalent constraints (`ObjectPointer`, `BuilderPointer`) and then constrain builders via `common.Builder[O, SO]` (as shown above).

### Scheme attacher correctness

To reduce false positives/negatives around scheme attachment, the helper performs a preflight:

- Create a fresh `runtime.Scheme`
- Run `SchemeAttacher(scheme)`
- Assert `scheme.ObjectKinds(testObject)` includes `ExpectedGVK`

This validates the test configuration and helps catch cases where the expected GVK or scheme attacher don't actually match the test object type.

### Parallelism

The helper **does not call** `t.Parallel()` internally. Callers can opt in at the test function level (as in the usage example) without the helper surprising them by running subtests in parallel.

## Summary

This design provides:

- **Reusable test helpers** that validate thin Pull wrappers
- **Black-box testing** that catches integration mistakes (argument swaps, wrong scheme attacher, wrong GVK)
- **Full generics** over builder, object, and API client type (`C runtimeclient.Client`)
- **Ergonomic call sites** via a type-inference-friendly constructor (`NewNamespacedPullTestConfig`)
- **Scheme-attacher/GVK preflight** using `runtime.Scheme.ObjectKinds`
- **Simple meta-testing** using `t.Run`’s boolean return value
