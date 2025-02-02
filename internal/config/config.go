package config

import (
    "fmt"

    "github.com/spf13/viper"
)

type Config struct {
    MakeMkvPath     string
    DiscDbDefs      string
    DriveNum        int
    MkvDest         string
    Disc            int
    Slug            string
}

func Load() (Config, error) {
    cfgFile := viper.GetString("config")
    fmt.Printf("%s", cfgFile)
    viper.SetConfigFile(cfgFile)
    err := viper.ReadInConfig()
    if err != nil {
        fmt.Printf("%s", err)
    }
    for key, setting := range viper.AllSettings() {
        fmt.Printf("%s: %v\n", key, setting)
    }
    
    var config Config
    err = viper.Unmarshal(&config)
    if err != nil {
        fmt.Printf("%s", err)
    }

    return config, nil
}
