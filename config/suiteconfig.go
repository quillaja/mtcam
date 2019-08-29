package config

// SuiteConfig contains settings shared among all executables
// in the cmd folder.
type SuiteConfig struct {
	DatabaseDriverName string
	DatabaseConnection string

	ImageRoot string
}
