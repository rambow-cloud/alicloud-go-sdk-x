# Implement bounded retry and full-jitter backoff

- GitHub issue: #6.

- Implement a replaceable policy, bounded attempts and per-policy retry budget; support Retry-After/context.
- Keep retries off by default and require idempotent replayable operations.
- Test independently, then integrate under #3.

- Acceptance includes implementation, meaningful offline behavior tests, English Go comments, executable Examples and equivalent bilingual guides.
- GitHub remains the source of execution status.
