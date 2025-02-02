package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
    MakeMkvPath     string
    DiscDbDefs      string
    MkvDest         string
    Disc            int
    Slug            string
}

func  main() {
    pflag.Int("disc", 1, "The number of the disc as defined in discdb.")
    pflag.String("slug", "", "The path to the slug of the disc.")

    pflag.Parse()

    viper.SetConfigName("config")
    viper.AddConfigPath("./")
    viper.BindPFlags(pflag.CommandLine)
    err := viper.ReadInConfig()
    if err != nil {
        fmt.Printf("%s", err)
    }
    
    var config Config
    err = viper.Unmarshal(&config)
    if err != nil {
        fmt.Printf("%s", err)
    }

    for _, key := range viper.AllKeys() {
        fmt.Printf("%s: %s\n", key, viper.Get(key))
    }

    fmt.Printf("%s", config)

    // Do something with this result
    checkPath(config.MakeMkvPath)
    checkPath(config.DiscDbDefs)
    
    // It might be useful to write the title results to a file that can be read back later.
    // Could allow for me to manually create a title map that this code reads in and just does the mapping if the disc wasn't ripped by this code
    // titles, err := readTitles(config) 
    // if err != nil {
    //     fmt.Printf("%s", err)
    // }
    // fmt.Printf("Titles: %s\n", titles)

    // Maybe eventually allow for discs to be chained, so that the code automatically preps for the next disc to be inserted and rips that one next?
    discMap, err := loadDiscDbDef(config)
    if err != nil {
        fmt.Printf("%s", err)
    }
    fmt.Printf("%s\n", discMap)
    // os.MkdirAll(config.MkvDest + "Oppenheimer", os.ModePerm)

    // Now inspect the files in the output folder and map them to their proper names in the disc db
}

type SummaryTitle struct {
    Name                    string
    SourceFileName          string
    Duration                string
    ChaptersCount           string
    Size                    string
    SegmentCount            string
    SegmentMap              string
    Type                    string
    Season                  string
    Episode                 string
    FileName                string
}

func checkPath(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true;
    }

    return false;
}
