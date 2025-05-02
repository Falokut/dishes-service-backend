package repository

type Secret struct {
	secret string
}

func NewSecret(secret string) Secret {
	return Secret{
		secret: secret,
	}
}

func (s Secret) GetSecret() string {
	return s.secret
}
