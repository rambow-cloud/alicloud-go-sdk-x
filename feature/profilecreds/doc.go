// Package profilecreds loads native Alibaba Cloud CLI JSON profiles and supplies
// cached temporary credentials without invoking the CLI. NewProvider supports
// OAuth, StsToken, AK, RamRoleArn, ChainableRamRoleArn, OIDC, EcsRamRole,
// CredentialsURI, External and CloudSSO. AK sources require explicit
// Options.AllowLongLived opt-in. Other modes return ErrUnsupportedMode.
//
// OAuth reuses valid STS state, refreshes access tokens and exchanges them for STS
// credentials using the pinned CLI protocol. Token rotation and STS state are
// atomically merged into the selected CLI profile, preserving all other settings.
// SDK writers use bounded file locks; concurrent CLI reconfiguration is unsupported.
// Initial interactive
// login and revoked sessions require aliyun configure --mode OAuth. Config files
// supply immutable configuration settings; OAuth session fields reload under lock.
// Credential HTTP work is lazy,
// bounded and shared; a canceled caller does not cancel another caller's refresh.
// Default formatting hides tokens, and errors retain errors.Is/As identity.
// CloudSSO reloads login tokens and reuses native STS state, then exchanges through
// the portal. It does not refresh or write CloudSSO login tokens; expired sessions
// require CLI re-login. URI/process/metadata retrieval uses bounded temporary sources.
package profilecreds
