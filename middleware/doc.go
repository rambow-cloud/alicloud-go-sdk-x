// Package middleware provides ordered operation and attempt interceptors.
// Initialize and Build execute once per operation; Finalize and Deserialize
// execute on each HTTP attempt. Middleware must be safe for concurrent use and
// call next at most once. Exchange belongs to one operation and must not escape
// Handle. Registrations are copied by NewStack; individual middleware remain shared.
package middleware
