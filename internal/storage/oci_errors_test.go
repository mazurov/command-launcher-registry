package storage

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOCIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		category string
		op       string
		err      error
		expected string
	}{
		{
			name:     "auth error during push",
			category: OCICategoryAuth,
			op:       OCIOpPush,
			err:      errors.New("token expired"),
			expected: "OCI authentication error during push: token expired",
		},
		{
			name:     "network error during pull",
			category: OCICategoryNetwork,
			op:       OCIOpPull,
			err:      errors.New("connection refused"),
			expected: "OCI network error during pull: connection refused",
		},
		{
			name:     "storage error during connect",
			category: OCICategoryStorage,
			op:       OCIOpConnect,
			err:      errors.New("registry unavailable"),
			expected: "OCI storage error during connect: registry unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ociErr := &OCIError{
				Category: tt.category,
				Op:       tt.op,
				Err:      tt.err,
			}
			assert.Equal(t, tt.expected, ociErr.Error())
		})
	}
}

func TestOCIError_Unwrap(t *testing.T) {
	underlying := errors.New("original error")
	ociErr := &OCIError{
		Category: OCICategoryAuth,
		Op:       OCIOpPush,
		Err:      underlying,
	}

	assert.Equal(t, underlying, ociErr.Unwrap())
	assert.True(t, errors.Is(ociErr, underlying))
}

func TestOCIError_Is_StorageUnavailable(t *testing.T) {
	ociErr := NewOCIAuthError(OCIOpPush, errors.New("auth failed"))

	// OCIError should be "Is" ErrStorageUnavailable for compatibility
	assert.True(t, errors.Is(ociErr, ErrStorageUnavailable))
}

func TestNewOCIAuthError(t *testing.T) {
	err := errors.New("invalid token")
	ociErr := NewOCIAuthError(OCIOpPush, err)

	assert.Equal(t, OCICategoryAuth, ociErr.Category)
	assert.Equal(t, OCIOpPush, ociErr.Op)
	assert.Equal(t, err, ociErr.Err)
}

func TestNewOCINetworkError(t *testing.T) {
	err := errors.New("timeout")
	ociErr := NewOCINetworkError(OCIOpPull, err)

	assert.Equal(t, OCICategoryNetwork, ociErr.Category)
	assert.Equal(t, OCIOpPull, ociErr.Op)
	assert.Equal(t, err, ociErr.Err)
}

func TestNewOCIStorageError(t *testing.T) {
	err := errors.New("not found")
	ociErr := NewOCIStorageError(OCIOpConnect, err)

	assert.Equal(t, OCICategoryStorage, ociErr.Category)
	assert.Equal(t, OCIOpConnect, ociErr.Op)
	assert.Equal(t, err, ociErr.Err)
}

func TestCategorizeOCIError_NilError(t *testing.T) {
	result := CategorizeOCIError(OCIOpPush, nil)
	assert.Nil(t, result)
}

func TestCategorizeOCIError_AuthErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "401 status code",
			err:  errors.New("status 401: Unauthorized"),
		},
		{
			name: "403 status code",
			err:  errors.New("HTTP 403 Forbidden"),
		},
		{
			name: "UNAUTHORIZED response",
			err:  errors.New("UNAUTHORIZED: access denied"),
		},
		{
			name: "FORBIDDEN response",
			err:  errors.New("FORBIDDEN: insufficient permissions"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ociErr := CategorizeOCIError(OCIOpPush, tt.err)
			assert.NotNil(t, ociErr)
			assert.Equal(t, OCICategoryAuth, ociErr.Category)
		})
	}
}

func TestCategorizeOCIError_NetworkErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "DNS error",
			err:  &net.DNSError{Name: "registry.example.com", Err: "no such host"},
		},
		{
			name: "URL timeout error",
			err:  &url.Error{Op: "Get", URL: "https://registry.example.com", Err: &timeoutError{}},
		},
		{
			name: "URL connection error",
			err:  &url.Error{Op: "Get", URL: "https://registry.example.com", Err: errors.New("connection refused")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ociErr := CategorizeOCIError(OCIOpPull, tt.err)
			assert.NotNil(t, ociErr)
			assert.Equal(t, OCICategoryNetwork, ociErr.Category)
		})
	}
}

func TestCategorizeOCIError_StorageErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "404 not found",
			err:  errors.New("status: 404 NOT_FOUND"),
		},
		{
			name: "500 internal error",
			err:  errors.New("HTTP 500 Internal Server Error"),
		},
		{
			name: "503 service unavailable",
			err:  errors.New("status 503: Service Unavailable"),
		},
		{
			name: "NOT_FOUND response",
			err:  errors.New("NOT_FOUND: repository does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ociErr := CategorizeOCIError(OCIOpPull, tt.err)
			assert.NotNil(t, ociErr)
			assert.Equal(t, OCICategoryStorage, ociErr.Category)
		})
	}
}

func TestCategorizeOCIError_DefaultsToStorage(t *testing.T) {
	err := errors.New("some unknown error")
	ociErr := CategorizeOCIError(OCIOpPush, err)

	assert.NotNil(t, ociErr)
	assert.Equal(t, OCICategoryStorage, ociErr.Category)
	assert.Equal(t, OCIOpPush, ociErr.Op)
}

func TestContainsHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		status   int
		expected bool
	}{
		{
			name:     "plain status code",
			errStr:   "error 401",
			status:   401,
			expected: true,
		},
		{
			name:     "status with prefix",
			errStr:   "HTTP status 403",
			status:   403,
			expected: true,
		},
		{
			name:     "status with colon",
			errStr:   "status: 404",
			status:   404,
			expected: true,
		},
		{
			name:     "HTTP prefix",
			errStr:   "HTTP 500 Internal Server Error",
			status:   500,
			expected: true,
		},
		{
			name:     "no match",
			errStr:   "connection refused",
			status:   401,
			expected: false,
		},
		{
			name:     "partial match (different status)",
			errStr:   "HTTP 200 OK",
			status:   401,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsHTTPStatus(tt.errStr, tt.status)
			assert.Equal(t, tt.expected, result, "errStr=%q status=%d", tt.errStr, tt.status)
		})
	}
}

// timeoutError implements net.Error for testing
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestCategorizeOCIError_PreservesOperation(t *testing.T) {
	ops := []string{OCIOpPush, OCIOpPull, OCIOpConnect}

	for _, op := range ops {
		t.Run(fmt.Sprintf("op=%s", op), func(t *testing.T) {
			ociErr := CategorizeOCIError(op, errors.New("some error"))
			assert.Equal(t, op, ociErr.Op)
		})
	}
}
