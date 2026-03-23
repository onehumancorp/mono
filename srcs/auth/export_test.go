package auth

// GetSecretForTest exposes the unexported secret for testing purposes within the auth package.
func (s *Store) GetSecretForTest() []byte {
	return s.secret
}

// SetOIDCConfigForTest exposes setting the unexported OIDC configuration for testing purposes.
func (s *Store) SetOIDCConfigForTest(cfg OIDCConfig) {
	s.oidcCfg = cfg
}
