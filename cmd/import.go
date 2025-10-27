package cmd

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var importCmd = &cobra.Command {
	Use: "import",
	Short: "Imports The Disc DB database",
	Run: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) {
	logger, ok := cmd.Context().Value("LOGGER").(*zap.SugaredLogger)
	if !ok {
		panic("Failed to retrieve logger from context. Unable to continue.")
	}

	cfg, ok := cmd.Context().Value("GLOBAL").(config.Config)

	logger.Infoln(calculateDiscHash(logger, cfg.DiscDbDefs));	
}

func calculateDiscHash(logger *zap.SugaredLogger, defDir string) string {
	files, _ := filepath.Glob("/home/maddux/Documents/data/data/series/Black Sails (2014)/2018-complete-collection-blu-ray/disc01.txt")

	hash := md5.New()
	for _, file := range files {
		logger.Infoln("Opening file: %s", file)
		temp, _ := os.Open(file)
		
		scanner := bufio.NewScanner(temp)
		for scanner.Scan() {
			line := scanner.Text()
			if (!strings.HasPrefix(line, "HSH")) {
				continue
			}

			logger.Infoln(line)

			size, _ := strconv.ParseUint(strings.TrimSpace(strings.Split(line, ",")[3]), 10, 64)
			logger.Infoln(size)
			bs := make([]byte, 8)
			binary.LittleEndian.PutUint64(bs, size)

			hash.Write(bs)
		}

		temp.Close()
	}

	return strings.ToUpper(fmt.Sprintf("%x", hash.Sum(nil)))
}
