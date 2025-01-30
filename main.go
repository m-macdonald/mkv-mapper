package main

import (
	// "bufio"
	"bufio"
	"fmt"
	"io"
	"strings"

	// "io"
	"os"
	"os/exec"

	//	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
    MakeMkvPath     string
    DiscDbDefs      string
    MkvDest         string
}

func  main() {
    viper.SetConfigName("config")
    viper.AddConfigPath("./")
    err := viper.ReadInConfig()
    fmt.Printf("%s", err)

    var config Config
    err = viper.Unmarshal(&config)
    fmt.Printf("%s", err)
    // Do something with this result
    checkPath(config.MakeMkvPath)
    checkPath(config.DiscDbDefs)

    readTitles(config)

    // os.MkdirAll(config.MkvDest + "Oppenheimer", os.ModePerm)

    // cmd := exec.Command(config.MakeMkvPath, "mkv", "disc:0", "all", config.MkvDest)
    // cmd.Stderr = os.Stderr
    // cmd.Stdout = os.Stdout
    // if err = cmd.Run(); err != nil {
    //     fmt.Printf("%s", err)
    // }
    // _, err = cmd.StdoutPipe()
    // if err = cmd.Start(); err != nil {
    //     fmt.Printf("%s", err)
    //
    //     return
    // }
    //
    // go func(p io.ReadCloser) {
    //     reader := bufio.NewReader(pipe)
    //     line, err := reader.ReadString('\n')
    //     for err == nil {
    //         fmt.Println(line)
    //         line, err = reader.ReadString('\n')
    //     }
    // }(pipe)

    //
    // if err = cmd.Wait(); err != nil {
    //     // Log the error
    //     fmt.Printf("%s", err)
    // }

    // Now inspect the files in the output folder and map them to their proper names in the disc db
}

type Title struct {
    FileName        string
    MplsName        string
}

func readTitles(config Config) {
    cmd := exec.Command(config.MakeMkvPath, "info", "disc:0", "--robot")
    pipe, err := cmd.StdoutPipe()
    cmd.Start()
    // titles := make(map[string]string)
    
    go func(p io.ReadCloser) {
        reader := bufio.NewReader(p)
        // Might need OS-specific files that define constants for things like a linebreak
        line, err := reader.ReadString('\n')
        fmt.Printf("%s", line)
        for err == nil {
            if (!strings.HasPrefix(line, "TINFO")) {
                line, err = reader.ReadString('\n')
                // fmt.Printf(line)
                continue
            }
            params := strings.Split(line[6:], ",")
            // fmt.Printf("%s", params)
    //Might need to consider using the track number for uniqueness just while parsing these
            // 16 is the mpls name
            // 27 is the name of that makemkv will give the file
            if (params[1] == "16") {
                fmt.Printf("%s", params[3])
            } else if (params[1] == "27") {
                fmt.Printf("%s", params[3])
            }

            line, err = reader.ReadString('\n')
        }
    }(pipe)

    if err = cmd.Wait(); err != nil {
        fmt.Printf("%s", err)
    }
}

func checkPath(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true;
    }

    return false;
}
