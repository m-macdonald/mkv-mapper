package config

const (
	DiscRoot    = "discRoot"
	LogLevel    = "logLevel"
	MakeMkvPath = "makemkvPath"
	OutputDir   = "outputDir"
)

type Config struct {
	DiscRoot    string `mapstructure:"discRoot"`
	LogLevel    string `mapstructure:"logLevel"`
	MakeMkvPath string `mapstructure:"makemkvPath"`
	OutputDir   string `mapstructure:"outputDir"`
}
