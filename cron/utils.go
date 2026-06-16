package cron

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
)

// JobError is an error that carries a stack trace captured at the point the
// error was created — inside the job function, before it returned.
// Construct one with Err() so the cron runner can report the full trace.
type JobError struct {
	cause error
	stack []byte
}

func (e *JobError) Error() string { return e.cause.Error() }
func (e *JobError) Unwrap() error { return e.cause }

// Err wraps err with a stack trace captured at the call site.
// Use it inside a cron job to get traces that reach the failure site:
//
//	func myJob() error {
//	    if err := doWork(); err != nil {
//	        return cron.Err(err)
//	    }
//	    return nil
//	}
//
// Without Err(), the cron runner can only capture the stack trace of the
// goroutine that called the job, which is not very useful for debugging.
func Err(err error) error {
	if err == nil {
		return nil
	}
	return &JobError{cause: err, stack: debug.Stack()}
}

// safeRun calls fn() and converts any panic into a *JobError whose stack is
// captured inside the recover — at which point the job's goroutine frames are
// still present on the stack.
func safeRun(fn func() error) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = &JobError{
				cause: fmt.Errorf("%v", r),
				stack: debug.Stack(),
			}
		}
	}()
	return fn()
}

// jobName returns the fully-qualified function name for fn (e.g. "main.myJob").
// Used by trimStack to locate the job's frame in the raw stack output.
func jobName(fn func() error) string {
	pc := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown"
	}
	return f.Name()
}

// jobLocation returns the source file and line number of fn's entry point.
// This is the function definition line, not the specific line that errored,
// but it is enough to locate the job in non-panic situations where no
// internal stack is available.
func jobLocation(fn func() error) string {
	pc := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown"
	}
	file, line := f.FileLine(pc)
	return fmt.Sprintf("%s:%d", file, line)
}

// trimStack cuts raw debug.Stack() output at the job's own frame.
// All frames from the goroutine header up to and including the job function
// are kept — Pine/runtime frames above the job are useful context.
func trimStack(raw []byte, name string) string {
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) < 2 {
		return strings.TrimSpace(string(raw))
	}

	header := lines[0] // "goroutine N [running]:"

	// Locate the job's function line. Function lines have no leading tab;
	// file lines do. We match on the function name so the cut point is
	// always the job itself, regardless of what called it.
	jobIdx := -1
	for i := 1; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "\t") && strings.Contains(lines[i], name) {
			jobIdx = i
			break
		}
	}
	if jobIdx == -1 {
		// Job frame not found — return the full stack untouched.
		return strings.TrimSpace(string(raw))
	}

	// Include the job's file/line pair if present.
	end := jobIdx + 1
	if end < len(lines) && strings.HasPrefix(lines[end], "\t") {
		end++
	}

	var kept []string
	i := 1
	for i < end {
		funcLine := lines[i]
		if strings.HasPrefix(funcLine, "\t") {
			i++
			continue
		}
		var fileLine string
		if i+1 < end && strings.HasPrefix(lines[i+1], "\t") {
			fileLine = lines[i+1]
		}
		kept = append(kept, funcLine)
		if fileLine != "" {
			kept = append(kept, fileLine)
			i += 2
		} else {
			i++
		}
	}

	if len(kept) == 0 {
		return header + "\n(no user frames retained)"
	}
	return header + "\n" + strings.Join(kept, "\n")
}

// formatError builds a structured error report:
//
//	job: <source file:line of the job function>
//	error: <err.Error()>
//	stack: <trimmed stack (panics) or explicit message (plain errors)>
func formatError(fn func() error, err error) string {
	var stack string
	var je *JobError
	if errors.As(err, &je) {
		stack = trimStack(je.stack, jobName(fn))
	} else {
		stack = "no panic — stack trace not available. Use cron.Err(err) inside the job to capture a stack trace at the error site."
	}
	return fmt.Sprintf("job: %s\nerror: %s\nstack: %s", jobLocation(fn), err.Error(), stack)
}
