package retry

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Func is a function with return error type that will be executed and evaluated by Executor
type Func func() error

// FuncHTTP is a function with return httpResponse and error(i.e: http status code). The httpResponse status code will be executed and evaluated by Executor
type FuncHTTP func() (*http.Response, error)

// Executor executes a closure, inspect the error, and do retry if necessary
func Executor(fn Func) error {
	return ExecutorWithPolicyType(StandardPolicy, fn)
}

// ExecutorWithPolicyType executes a func, inspect the error and evaluate based on retryPolicies, and do retry if necessary
func ExecutorWithPolicyType(policyType PolicyType, fn Func) error {
	retryPolicies := GetRetryPolicies(policyType)
	return ExecutorWithPolicies(retryPolicies, fn)
}

// ExecutorWithPolicies executes a func, inspect the error and evaluate based retryPolicies, and do retry if necessary
func ExecutorWithPolicies(retryPolicies []Policy, fn Func) error {
	err := fn()
	if err != nil {
		var attempt = 1
		for {
			delay, limit, ok := shouldRetry(retryPolicies, 0, err.Error())
			if ok && attempt <= limit {
				time.Sleep(delay)
				err = fn()
				if err == nil {
					return nil
				}
				attempt++
			} else {
				return err
			}
		}
	}
	return err
}

// ExecutorHTTP executes a closure, inspect the error, and do retry if necessary
func ExecutorHTTP(fn FuncHTTP) error {
	return ExecutorHTTPWithPolicyType(StandardPolicy, fn)
}

// ExecutorHTTPWithPolicyType executes a func, inspect the error and evaluate based on retryPolicies, and do retry if necessary
func ExecutorHTTPWithPolicyType(policyType PolicyType, fn FuncHTTP) error {
	retryPolicies := GetRetryPolicies(policyType)
	return ExecutorHTTPWithPolicies(retryPolicies, fn)
}

// ExecutorHTTPWithPolicies executes a func, inspect the error and evaluate based retryPolicies, and do retry if necessary
func ExecutorHTTPWithPolicies(retryPolicies []Policy, fn FuncHTTP) error {
	resp, err := fn()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		var attempt = 1
		for {
			delay, limit, ok := shouldRetry(retryPolicies, int(resp.StatusCode), resp.Status)
			if ok && attempt <= limit {
				time.Sleep(delay)
				resp, err = fn()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return nil
				}
				attempt++
			} else {
				return fmt.Errorf("ERROR: httpStatusCode: %d, httpStatus: %s", resp.StatusCode, resp.Status)
			}
		}
	}
	return err
}

// GetRetryPolicies returns list of retry policies
func GetRetryPolicies(policyType PolicyType) []Policy {
	var policies []Policy
	switch policyType {
	case HTTPPolicy:
		policies = []Policy{
			{
				ErrorCodeNumber: http.StatusServiceUnavailable,
				ErrorCodeString: http.StatusText(http.StatusServiceUnavailable),
				DelayDuration:   time.Second * 2,
				RetryLimit:      3,
			},
			{
				ErrorCodeNumber: http.StatusRequestTimeout,
				ErrorCodeString: http.StatusText(http.StatusRequestTimeout),
				DelayDuration:   time.Second * 2,
				RetryLimit:      3,
			},
		}
	case StandardPolicy:
		policies = []Policy{
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
	}
	return policies
}

func shouldRetry(criteria []Policy, errCodeNumber int, errCodeString string) (time.Duration, int, bool) {
	if criteria == nil {
		return time.Duration(0), 0, false
	}
	for _, c := range criteria {

		if c.ErrorCodeNumber == errCodeNumber &&
			c.ErrorCodeString == errCodeString ||
			strings.Contains(strings.ToLower(errCodeString), strings.ToLower(c.ErrorCodeString)) {
			return c.DelayDuration, c.RetryLimit, true
		}
	}
	return time.Duration(0), 0, false
}

// Policy will be evaluated by Executor to determine if a certain error that's
// returned by certain operation can be retried
type Policy struct {
	ErrorCodeNumber int
	ErrorCodeString string
	DelayDuration   time.Duration
	RetryLimit      int
}

// PolicyType is an enum for list of retryable criteria
// This enum can be expanded as we have more types of execution
type PolicyType int

const (
	// HTTPPolicy criteria
	HTTPPolicy PolicyType = iota

	// StandardPolicy function call
	StandardPolicy
)
