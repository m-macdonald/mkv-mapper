package config

const (
	DiscRoot         = "discRoot"
	LogLevel         = "logLevel"
	MakeMkvPath      = "makemkvPath"
	OutputDir        = "outputDir"
	TemplateOverride = "templates.override"
)

type Config struct {
	DiscRoot    string         `mapstructure:"discRoot"`
	LogLevel    string         `mapstructure:"logLevel"`
	MakeMkvPath string         `mapstructure:"makemkvPath"`
	OutputDir   string         `mapstructure:"outputDir"`
	Templates   TemplateConfig `mapstructure:"templates"`
}

type TemplateConfig struct {
	Episode  string `mapstructure:"episode"`
	Extra    string `mapstructure:"extra"`
	Movie    string `mapstructure:"movie"`
	Override string `mapstructure:"override"`
	Unknown  string `mapstructure:"unknown"`
}
