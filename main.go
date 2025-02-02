/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
    "fmt"
	"os"

	"m-macdonald/mkv-mapper/cmd"
	"m-macdonald/mkv-mapper/internal/config"
)

func main() {
	cmd.Execute()
}

func realMain() int {
    config, err := config.Load()
    if err != nil {
        fmt.Printf("%s", err)

        return 1
    }


    return 0
}


