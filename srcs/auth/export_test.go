package auth

// Export internal functions and fields for testing
var ParseHS256ForTest = parseHS256
var SignHS256ForTest = signHS256

func (s *Store) GetSecretForTest() []byte {
	return s.secret
}

func (s *Store) SetOIDCConfigForTest(cfg OIDCConfig) {
	s.oidcCfg = cfg
}
