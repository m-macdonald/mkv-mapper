package config

const (
	DiscRoot     = "discRoot"
	FilenameTmpl = "filenameTmpl"
	LogLevel     = "logLevel"
	MakeMkvPath  = "makemkvPath"
	OutputDir    = "outputDir"
)

type Config struct {
	DiscRoot     string `mapstructure:"discRoot"`
	LogLevel     string `mapstructure:"logLevel"`
	MakeMkvPath  string `mapstructure:"makemkvPath"`
	OutputDir    string `mapstructure:"outputDir"`
	FilenameTmpl string `mapstructure:"filenameTmpl"`
}
