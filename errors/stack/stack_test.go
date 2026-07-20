package stack

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type testError struct {
	cause error
}

func (err *testError) Error() string { return "typed error" }
func (err *testError) Unwrap() error { return err.cause }

type testAppError struct {
	code    string
	message string
}

func (err *testAppError) Error() string { return err.message }

func TestWith_NilReturnsNil(t *testing.T) {
	if got := With(nil); got != nil {
		t.Fatalf("With(nil) = %v, want nil", got)
	}
}

func TestWith_PreservesErrorChain(t *testing.T) {
	cause := errors.New("cause")
	typed := &testError{cause: cause}
	wrapped := With(fmt.Errorf("context: %w", typed))

	if !errors.Is(wrapped, cause) {
		t.Fatal("errors.Is() did not find the cause")
	}
	got, ok := errors.AsType[*testError](wrapped)
	if !ok || got != typed {
		t.Fatalf("errors.AsType() = (%v, %v), want (%v, true)", got, ok, typed)
	}
	if _, ok := errors.AsType[Provider](wrapped); !ok {
		t.Fatal("errors.AsType() did not find Provider")
	}
}

func TestWith_IdempotentPreservesFirstStack(t *testing.T) {
	first := captureFirstStack()
	provider := stackProvider(t, first)
	firstStack := provider.Stack()

	tree := errors.Join(first, errors.New("other error"))
	if got := With(tree); got != tree {
		t.Fatal("With() replaced an error tree with a stack")
	}
	if got := stackProvider(t, tree).Stack(); got != firstStack {
		t.Fatalf("Stack() = %q, want first stack %q", got, firstStack)
	}
}

func TestWithSkip_OmitsHelperFrame(t *testing.T) {
	captured := stackProvider(t, captureWithSkippedHelper()).Stack()
	if !strings.Contains(captured, "TestWithSkip_OmitsHelperFrame") {
		t.Fatalf("Stack() = %q, want test caller frame", captured)
	}
	if strings.Contains(captured, "captureWithSkippedHelper") {
		t.Fatalf("Stack() = %q, contains skipped helper", captured)
	}
}

func TestUnwrap_RemovesOnlyTopLevelWrapper(t *testing.T) {
	cause := errors.New("cause")
	wrapped := With(cause)
	if got := Unwrap(wrapped); got != cause {
		t.Fatalf("Unwrap() = %v, want %v", got, cause)
	}

	outer := fmt.Errorf("outer context: %w", wrapped)
	if got := Unwrap(outer); got != outer {
		t.Fatal("Unwrap() removed a nested stack")
	}
	if !errors.Is(outer, cause) {
		t.Fatal("outer error no longer preserves the cause")
	}
}

func TestProvider_StackIsStableAcrossCalls(t *testing.T) {
	provider := stackProvider(t, With(errors.New("cause")))
	first := provider.Stack()
	if first == "" {
		t.Fatal("Stack() returned an empty stack")
	}
	if strings.Contains(first, "\n") {
		t.Fatalf("Stack() = %q, want a single-line stack", first)
	}
	if got := provider.Stack(); got != first {
		t.Fatalf("second Stack() = %q, want %q", got, first)
	}
}

func TestCapture_ReturnsCallerFrame(t *testing.T) {
	captured := captureFallbackStack()
	if !strings.Contains(captured, "captureFallbackStack") {
		t.Fatalf("Capture() = %q, want caller frame", captured)
	}
}

func TestWith_BoundsCapturedProgramCounters(t *testing.T) {
	wrapped := captureDeepStack(maxFrames + 8)
	stack, ok := wrapped.(*stackError)
	if !ok {
		t.Fatalf("With() returned %T, want *stackError", wrapped)
	}
	if len(stack.pcs) > maxFrames {
		t.Fatalf("captured %d program counters, want at most %d", len(stack.pcs), maxFrames)
	}
}

func TestReplace_ReplacesDisplayAndPreservesErrorTree(t *testing.T) {
	cause := &testError{cause: errors.New("record not found")}
	appErr := &testAppError{code: "USER_NOT_FOUND", message: "user not found"}
	replaced := Replace(cause, appErr)

	if got := replaced.Error(); got != appErr.message {
		t.Fatalf("Replace().Error() = %q, want %q", got, appErr.message)
	}
	gotAppErr, ok := errors.AsType[*testAppError](replaced)
	if !ok || gotAppErr != appErr || gotAppErr.code != "USER_NOT_FOUND" {
		t.Fatalf("errors.AsType() = (%v, %v), want AppError USER_NOT_FOUND", gotAppErr, ok)
	}
	if !errors.Is(replaced, cause) {
		t.Fatal("errors.Is() did not find the original cause")
	}
	if !errors.Is(replaced, appErr) {
		t.Fatal("errors.Is() did not find the replacement")
	}
}

func TestReplace_PrefersOuterTypedErrorAndPreservesInnerStack(t *testing.T) {
	inner := &testAppError{code: "REPOSITORY_NOT_FOUND", message: "record not found"}
	captured := With(inner)
	innerStack := stackProvider(t, captured)
	outer := &testAppError{code: "USER_NOT_FOUND", message: "user not found"}
	replaced := Replace(captured, outer)

	got, ok := errors.AsType[*testAppError](replaced)
	if !ok || got != outer {
		t.Fatalf("errors.AsType() = (%v, %v), want outer %v", got, ok, outer)
	}
	if !errors.Is(replaced, outer) {
		t.Fatal("errors.Is() did not find the outer replacement")
	}
	if !errors.Is(replaced, inner) {
		t.Fatal("errors.Is() did not find the inner cause")
	}
	if got := stackProvider(t, replaced); got != innerStack {
		t.Fatalf("errors.AsType[Provider]() = %T, want inner stack provider", got)
	}
}

func TestReplace_PreservesCapturedStack(t *testing.T) {
	captured := captureFirstStack()
	firstStack := stackProvider(t, captured).Stack()
	replaced := Replace(captured, &testAppError{code: "USER_NOT_FOUND", message: "user not found"})

	if got := stackProvider(t, replaced).Stack(); got != firstStack {
		t.Fatalf("Stack() = %q, want %q", got, firstStack)
	}
	if got := With(replaced); got != replaced {
		t.Fatal("With() added a second stack after Replace()")
	}
	if got := Unwrap(replaced); got != replaced {
		t.Fatal("Unwrap() removed a nested stack after Replace()")
	}
}

func TestReplace_CapturesCallerStackWhenCauseHasNone(t *testing.T) {
	replaced := captureReplacementStack()
	firstFrame := strings.SplitN(stackProvider(t, replaced).Stack(), "; ", 2)[0]

	if !strings.Contains(firstFrame, "captureReplacementStack") {
		t.Fatalf("first stack frame = %q, want Replace caller", firstFrame)
	}
	if strings.Contains(firstFrame, ".Replace ") || strings.Contains(firstFrame, ".with ") {
		t.Fatalf("first stack frame = %q, want no internal stack frame", firstFrame)
	}
}

func TestReplace_MultipleCausesMapToSameAppErrorCode(t *testing.T) {
	causes := []error{errors.New("record not found"), errors.New("soft deleted")}
	for _, cause := range causes {
		replaced := Replace(cause, &testAppError{code: "USER_NOT_FOUND", message: "user not found"})
		appErr, ok := errors.AsType[*testAppError](replaced)
		if !ok || appErr.code != "USER_NOT_FOUND" {
			t.Fatalf("errors.AsType() = (%v, %v), want USER_NOT_FOUND", appErr, ok)
		}
		if !errors.Is(replaced, cause) {
			t.Fatalf("errors.Is() did not find cause %v", cause)
		}
	}
}

func TestReplace_NilHandling(t *testing.T) {
	appErr := &testAppError{code: "USER_NOT_FOUND", message: "user not found"}
	if got := Replace(nil, appErr); got != nil {
		t.Fatalf("Replace(nil, appErr) = %v, want nil", got)
	}
	cause := errors.New("cause")
	if got := Replace(cause, nil); got != cause {
		t.Fatalf("Replace(cause, nil) = %v, want %v", got, cause)
	}
}

func captureFirstStack() error {
	return With(errors.New("first error"))
}

func captureReplacementStack() error {
	return Replace(errors.New("cause"), &testAppError{code: "USER_NOT_FOUND", message: "user not found"})
}

func captureWithSkippedHelper() error {
	return WithSkip(errors.New("cause"), 1)
}

func captureFallbackStack() string {
	return Capture(0)
}

func captureDeepStack(depth int) error {
	if depth == 0 {
		return With(errors.New("deep error"))
	}
	return captureDeepStack(depth - 1)
}

func stackProvider(t *testing.T, err error) Provider {
	t.Helper()
	provider, ok := errors.AsType[Provider](err)
	if !ok {
		t.Fatalf("errors.AsType[Provider](%T) = false", err)
	}
	return provider
}
