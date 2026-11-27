package model

// IdentityReference identifies a user in an external identity system.
// Issuer and Subject together distinguish identities from different providers.
// Authentication and provider integration belong outside this package.
type IdentityReference struct {
	Issuer  string
	Subject string
}
