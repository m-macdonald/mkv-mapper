package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	CachePath           = "cachePath"
	ConfigFilename      = "config"
	DiscBackupOutputDir = "disc.backup.outputDir"
	DiscRipBackup       = "disc.rip.backup"
	DiscRipBackupKeep   = "disc.rip.keepBackup"
	DiscMode            = "disc.mode"
	DiscRoot            = "discRoot"
	EnvPrefix           = "MKVMAP"
	LogLevel            = "logLevel"
	MakeMkvPath         = "makemkvPath"
	OutputDir           = "outputDir"
	ProgramDirname      = "mkv-mapper"
	TemplateOverride    = "templates.override"
)

type Config struct {
	CachePath   string         `mapstructure:"cachePath"`
	DiscRoot    string         `mapstructure:"discRoot"`
	LogLevel    string         `mapstructure:"logLevel"`
	MakeMkvPath string         `mapstructure:"makemkvPath"`
	OutputDir   string         `mapstructure:"outputDir"`
	Templates   TemplateConfig `mapstructure:"templates"`
	Disc        DiscConfig     `mapstructure:"disc"`
}

type DiscConfig struct {
	Mode   SelectionMode `mapstructure:"mode"`
	Backup BackupConfig  `mapstructure:"backup"`
	Rip    RipConfig     `mapstructure:"rip"`
}

type RipConfig struct {
	Backup     bool `mapstructure:"backup"`
	KeepBackup bool `mapstructure:"keepBackup"`
}

type BackupConfig struct {
	OutputDir string `mapstructure:"outputDir"`
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
		DiscRoot: "",
		LogLevel: "info",
		// Assume makemkvcon is on the path
		MakeMkvPath: "makemkvcon",
		OutputDir: "~/Videos/{{.Disc.Label}}",
		Templates: TemplateConfig{
			Movie:   "{{.Media.Title}} ({{.Disc.Year}})",
			Episode: "{{.Media.Title}}/Season {{.Item.Season}}/{{.Disc.SeriesTitle}} - S{{pad 2 .Item.Season}}E{{.Item.Episode}} - {{.Item.Title}}",
			Extra:   "Extras/{{.Item.Title}}",
			Unknown: "{{.MakeMkv.OutputFileName}}",
		},
		Disc: DiscConfig{
			Mode: ModeFullAuto,
			Backup: BackupConfig{
				OutputDir: "~/Videos/backup/{{.Disc.Label}}",
			},
		},
	}
}

func Load() (Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	merged := mergeConfig(DefaultConfig(), cfg)
	finalized, err := finalizeConfig(merged)
	if err != nil {
		return Config{}, err
	}

	return finalized, nil
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
	result.Disc = mergeDisc(base.Disc, user.Disc)

	return result
}

func mergeDisc(base DiscConfig, user DiscConfig) DiscConfig {
	result := base

	if user.Mode != "" {
		result.Mode = user.Mode
	}
	result.Backup = mergeBackup(base.Backup, user.Backup)
	result.Rip = mergeRip(base.Rip, user.Rip)

	return result
}

func mergeRip(base RipConfig, user RipConfig) RipConfig {
	result := base

	result.Backup = user.Backup
	result.KeepBackup = user.KeepBackup

	return result
}

func mergeBackup(base BackupConfig, user BackupConfig) BackupConfig {
	result := base

	if user.OutputDir != "" {
		result.OutputDir = user.OutputDir
	}

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
		result.CachePath = filepath.Join(cacheDir, ProgramDirname, "cache.sqlite")
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

	result.Disc.Backup.OutputDir, err = resolveAbsPath(result.Disc.Backup.OutputDir)
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

// expands environment variables in the path, and ~ to the user's home directory.
// This is needed for paths read from config files where
// we don't benefit from shell expansion
func resolveHomePath(path string) (string, error) {
	path = os.ExpandEnv(path)

	if strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}
