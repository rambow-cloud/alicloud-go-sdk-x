// Package sdktest provides offline transports and a deterministic clock for SDK
// tests. Script exhaustion always fails instead of falling back to the network.
// A scripted transport consumes steps under a mutex; checks can run concurrently
// and must be concurrency safe. Clock.Sleep advances virtual time immediately;
// it does not advance Go's context timers. No real credentials are required.
package sdktest
