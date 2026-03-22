package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	DiscRoot         = "discRoot"
	LogLevel         = "logLevel"
	MakeMkvPath      = "makemkvPath"
	OutputDir        = "outputDir"
	TemplateOverride = "templates.override"
	CachePath        = "cachePath"
	appPath          = "mkv-mapper"
)

type Config struct {
	CachePath   string         `mapstructure:"cachePath"`
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

func DefaultConfig() Config {
	return Config{
		CachePath: "",
		// TODO: Eventually handle this by OS. For now it's best that the user just override
		DiscRoot:    "",
		LogLevel:    "info",
		// Assume makemkvcon is on the path
		MakeMkvPath: "makemkvcon",
		// Output to CWD by default
		OutputDir: ".",
		Templates: TemplateConfig{
			Movie:   "{{.Media.Title}} ({{.Disc.Year}})",
			Episode: "{{.Media.Title}}/Season {{.Item.Season}}/{{.Disc.SeriesTitle}} - S{{pad 2 .Item.Season}}E{{.Item.Episode}} - {{.Item.Title}}",
			Extra:   "Extras/{{.Item.Title}}",
			Unknown: "{{.MakeMkv.OutputFileName}}",
		},
	}
}

func ResolveConfig(base Config, user Config) (Config, error) {
	merged := mergeConfig(base, user)
	resolved, err := finalizeConfig(merged)
	if err != nil {
		return Config{}, err
	}

	return resolved, nil
}

func mergeConfig(base Config, user Config) Config {
	result := base

	if user.CachePath != "" {
		result.CachePath = user.CachePath
	}

	if user.DiscRoot != "" {
		result.DiscRoot = user.DiscRoot
	}

	if user.LogLevel != "" {
		result.LogLevel = user.LogLevel
	}

	if user.MakeMkvPath != "" {
		result.MakeMkvPath = user.MakeMkvPath
	}

	if user.OutputDir != "" {
		result.OutputDir = user.OutputDir
	}

	result.Templates = mergeTemplates(base.Templates, user.Templates)

	return result
}

func mergeTemplates(base TemplateConfig, user TemplateConfig) TemplateConfig {
	result := base

	if user.Movie != "" {
		result.Movie = user.Movie
	}

	if user.Episode != "" {
		result.Episode = user.Episode
	}

	if user.Extra != "" {
		result.Extra = user.Extra
	}

	if user.Unknown != "" {
		result.Unknown = user.Unknown
	}

	if user.Override != "" {
		result.Override = user.Override
	}

	return result
}

func finalizeConfig(config Config) (Config, error) {
	result := config

	if config.CachePath == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return Config{}, err
		}
		result.CachePath = filepath.Join(cacheDir, appPath, "cache.sqlite")
	}
	var err error
	result.CachePath, err = resolveAbsPath(result.CachePath)
	if err != nil {
		return Config{}, err
	}

	result.DiscRoot, err = resolveAbsPath(result.DiscRoot)
	if err != nil {
		return Config{}, err
	}

	result.MakeMkvPath, err = resolveExecutable(result.MakeMkvPath)
	if err != nil {
		return Config{}, err
	}

	result.OutputDir, err = resolveAbsPath(result.OutputDir)
	if err != nil {
		return Config{}, err
	}

	return result, nil
}

func resolveAbsPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	path, err := resolveHomePath(path)
	if err != nil {
		return "", err
	}

	return filepath.Abs(path)
}

func resolveExecutable(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	path, err := resolveHomePath(value)
	if err != nil {
		return "", err 
	}

	path, err = exec.LookPath(path)
	if err != nil {
		return "", fmt.Errorf("resolve executable %q: %w", value, err)
	}

	return path, nil
}

// TODO: Pretty sure this falls apart on Windows
func resolveHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}
