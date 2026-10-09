// Package externalcreds retrieves temporary credentials from ECS IMDSv2,
// explicit credential URIs, explicit processes and native CloudSSO sessions.
// Constructors perform no network or process work. Wrap providers in
// credentials.Cache for bounded coalesced renewal. Providers are immutable;
// transports and token callbacks remain shared and must support concurrency.
// Default diagnostics and formatting hide URLs, command arguments and secrets.
// HTTP/process work has a total timeout and bounded output; no automatic retry,
// redirect, interactive login or IMDSv1 downgrade occurs. Native profile loading
// and discovery precedence are supplied by config and feature/profilecreds.
package externalcreds
