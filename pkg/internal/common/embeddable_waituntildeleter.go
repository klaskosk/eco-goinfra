package common

import (
	"context"
	"time"
)

// EmbeddableWaitUntilDeleter is a mixin which provides the WaitUntilDeleted method to the embedding builder.
// The list type parameter L must correspond to the object type O.
type EmbeddableWaitUntilDeleter[O, L any, SO ObjectPointer[O], SL ListPointer[L]] struct {
	base Builder[O, SO]
}

// SetBase sets the base builder for the mixin. When the WaitUntilDeleted method is called, the common WaitUntilDeleted
// method will be called on the base builder. In practice, this can be either the EmbeddableBuilder or the
// resource-specific builder.
func (waiter *EmbeddableWaitUntilDeleter[O, L, SO, SL]) SetBase(base Builder[O, SO]) {
	waiter.base = base
}

// WaitUntilDeleted watches the resource until it is deleted or the timeout expires.
func (waiter *EmbeddableWaitUntilDeleter[O, L, SO, SL]) WaitUntilDeleted(timeout time.Duration) error {
	return WaitUntilDeleted[O, L, SO, SL](context.TODO(), waiter.base, timeout)
}

// EmbeddableDeleteAndWaiter is a mixin which provides the DeleteAndWait method to the embedding builder.
// The list type parameter L must correspond to the object type O.
type EmbeddableDeleteAndWaiter[O, L any, SO ObjectPointer[O], SL ListPointer[L]] struct {
	base Builder[O, SO]
}

// SetBase sets the base builder for the mixin. When the DeleteAndWait method is called, the common DeleteAndWait method
// will be called on the base builder. In practice, this can be either the EmbeddableBuilder or the resource-specific
// builder.
func (deleter *EmbeddableDeleteAndWaiter[O, L, SO, SL]) SetBase(base Builder[O, SO]) {
	deleter.base = base
}

// DeleteAndWait deletes the resource and waits for it to be fully removed from the cluster.
func (deleter *EmbeddableDeleteAndWaiter[O, L, SO, SL]) DeleteAndWait(timeout time.Duration) error {
	return DeleteAndWait[O, L, SO, SL](context.TODO(), deleter.base, timeout)
}

// EmbeddableDeleteAndWaitReturner is a mixin which provides the DeleteAndWait method to the embedding builder. The
// DeleteAndWait method returns the builder and the error from the DeleteAndWait method. To maintain compatibility with
// existing DeleteAndWait methods which return the builder, this struct has more complicated type parameters than the
// EmbeddableDeleteAndWaiter.
//
// Consumers of this mixin should set the base to the embedding builder rather than the EmbeddableBuilder so that
// DeleteAndWait returns the correct type.
type EmbeddableDeleteAndWaitReturner[O, L, B any, SO ObjectPointer[O], SL ListPointer[L], SB BuilderPointer[B, O, SO]] struct {
	base SB
}

// SetBase sets the base builder for the mixin. When the DeleteAndWait method is called, the common DeleteAndWait method
// will be called on the base builder. For EmbeddableDeleteAndWaitReturner, the base should be the resource-specific
// builder rather than EmbeddableBuilder.
func (deleter *EmbeddableDeleteAndWaitReturner[O, L, B, SO, SL, SB]) SetBase(base SB) {
	deleter.base = base
}

// DeleteAndWait deletes the resource and waits for it to be fully removed from the cluster. Regardless of the error,
// the builder is returned.
func (deleter *EmbeddableDeleteAndWaitReturner[O, L, B, SO, SL, SB]) DeleteAndWait(timeout time.Duration) (SB, error) {
	return deleter.base, DeleteAndWait[O, L, SO, SL](context.TODO(), deleter.base, timeout)
}
