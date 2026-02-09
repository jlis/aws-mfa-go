package app

import "os"

// OSEnv is the production environment provider.
type OSEnv struct{}

func (OSEnv) Get(key string) string { return os.Getenv(key) }
