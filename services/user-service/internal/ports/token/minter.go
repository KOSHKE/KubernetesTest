package token

// Subject describes the minimal subject data required to mint a token
type Subject interface {
	GetID() string
	GetEmail() string
}

// Minter is an abstraction for an access-token minter
type Minter interface {
	GenerateAccessToken(subject Subject) (string, error)
}
