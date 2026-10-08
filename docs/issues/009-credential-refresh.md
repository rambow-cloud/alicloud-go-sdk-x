# Implement expiry-aware shared credential caching

- GitHub issue: #9.

- Extend credentials with expiration; coalesce bounded refreshes, refresh early and never serve expired values.
- One canceled waiter must not cancel other waiters.
- Test with an injected clock/source and integrate the STS provider separately.

- Acceptance includes implementation, meaningful offline behavior tests, English Go comments, executable Examples and equivalent bilingual guides.
- GitHub remains the source of execution status.
