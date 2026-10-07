// Package retry provides opt-in idempotent retry policy, a bounded shared retry
// budget and cancellation-aware backoff. Standard uses full jitter and at most
// three attempts by default. Runtime retries are disabled unless configured.
// Custom Retryer implementations must be concurrency safe; ShouldRetry may
// consume budget and must be called once per failed attempt.
package retry
