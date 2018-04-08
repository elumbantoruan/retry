package retry

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutor(t *testing.T) {
	err := Executor(func() error {
		_, err := testOne()
		return err
	})
	assert.Equal(t, true, err == nil)
}

func TestExecutorWithPoliciesForTimedOutRecovered(t *testing.T) {
	policies := []Policy{
		{
			ErrorCodeString: "timedout",
			DelayDuration:   time.Second * 2,
			RetryLimit:      3,
		},
		{
			ErrorCodeString: "timed out",
			DelayDuration:   time.Second * 2,
			RetryLimit:      3,
		},
	}
	// will throw exception 1 time and it will be retried and it could be recovered
	// so err should be nil
	indexTestTimedout = 1
	err := ExecutorWithPolicies(policies, func() error {
		return testTimedout(1)
	})
	assert.Equal(t, true, err == nil)
}

func TestExecutorWithPolicyForTimedOutRecovered(t *testing.T) {
	// will throw exception 1 time and it will be retried and it could be recovered
	// so err should be nil
	indexTestTimedout = 1
	err := ExecutorWithPolicyType(StandardPolicy, func() error {
		err := testTimedout(1)
		return err
	})
	assert.Equal(t, true, err == nil)
}

func TestExecutorWithPoliciesForTimedOutNotRecovered(t *testing.T) {
	policies := []Policy{
		{
			ErrorCodeString: "timedout",
			DelayDuration:   time.Millisecond * 100,
			RetryLimit:      3,
		},
		{
			ErrorCodeString: "timed out",
			DelayDuration:   time.Millisecond * 100,
			RetryLimit:      3,
		},
	}
	// will throw exception 5 times and it will be retried but it will be not recovered
	// because the number of retry attempts exceeds the retry limit of 3
	// so err should not be nil
	indexTestTimedout = 1
	err := ExecutorWithPolicies(policies, func() error {
		return testTimedout(5)
	})
	assert.Equal(t, true, err != nil)
}

func TestExecutorWithPolicyTypeForTimedOutNotRecovered(t *testing.T) {
	// will throw exception 5 times and it will be retried but it will be not recovered
	// because the number of retry attempts exceeds the retry limit of 3
	// so err should not be nil
	indexTestTimedout = 1
	err := ExecutorWithPolicyType(StandardPolicy, func() error {
		return testTimedout(5)
	})
	assert.Equal(t, true, err != nil)
}

func TestExecutorWithPoliciesForNonRetriableError(t *testing.T) {
	policies := []Policy{
		{
			ErrorCodeString: "timedout",
			DelayDuration:   time.Millisecond * 100,
			RetryLimit:      3,
		},
		{
			ErrorCodeString: "timed out",
			DelayDuration:   time.Millisecond * 100,
			RetryLimit:      3,
		},
	}
	// will throw exception 5 times and it will be retried but it will be not recovered
	// because the number of retry attempts exceeds the retry limit of 3
	// so err should not be nil
	indexTestTimedout = 1
	err := ExecutorWithPolicies(policies, func() error {
		return testNonRetryableError(5)
	})
	assert.Equal(t, true, err != nil)
}

func TestExecutorWithPolicyForHTTPRecovered(t *testing.T) {
	// will throw exception 1 time and it will be retried and it could be recovered
	// so err should be nil
	indexTestTimedout = 1
	err := ExecutorHTTPWithPolicyType(HTTPPolicy, func() (*http.Response, error) {
		return testHTTPRetryable(1)
	})
	assert.Equal(t, true, err == nil)
}

func TestExecutorWithPolicyForHTTPNotRecovered(t *testing.T) {
	// will throw exception 4 time and it will be retried but it could not be recovered
	// as the number of error exceeds number of retry limit
	indexTestTimedout = 1
	err := ExecutorHTTPWithPolicyType(HTTPPolicy, func() (*http.Response, error) {
		return testHTTPRetryable(4)
	})
	assert.Equal(t, true, err != nil)
}

func testOne() (string, error) {
	return "test", nil
}

var indexTestTimedout = 0

func testTimedout(n int) error {
	if indexTestTimedout <= n {
		indexTestTimedout++
		return errors.New("timed out")
	}
	return nil
}

func testNonRetryableError(n int) error {
	if indexTestTimedout <= n {
		indexTestTimedout++
		return errors.New("something else")
	}
	return nil
}

func testHTTPRetryable(n int) (*http.Response, error) {
	if indexTestTimedout <= n {
		indexTestTimedout++
		resp := http.Response{
			StatusCode: http.StatusRequestTimeout,
			Status:     http.StatusText(http.StatusRequestTimeout),
		}
		return &resp, nil
	}
	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
	}
	return &resp, nil
}
