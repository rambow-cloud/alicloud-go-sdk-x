// Package waiter provides typed, bounded polling with explicit acceptors.
// A total wait deadline includes each fetch, its operation retries and sleeps.
// Caller cancellation remains distinguishable from waiter expiration. Waiters
// can be shared only when fetchers, acceptors and injected clocks are concurrency
// safe. Service adapters must review state rules; empty results never imply success.
package waiter
