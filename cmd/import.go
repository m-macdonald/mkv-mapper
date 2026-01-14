package cmd

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"os"
	"path/filepath"
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

	logger.Infoln(calculateDiscHash(logger, cfg.DiscDbDefs))
	logger.Infoln(cfg)

	discdb.Index(logger)
}

// func findDiscByHash(logger *zap.SugaredLogger, defDir string) {
// 	// hash := "32E7871A9A7170B6DA00CE548E65E925"
// 	logger.Infof("Definition Directory: %s", defDir)
//
// 	// discdb.FindDiscByHash(logger, defDir, hash)
// 	discdb.Index(logger)
// }

func calculateDiscHash(logger *zap.SugaredLogger, defDir string) string {
	files, _ := filepath.Glob("/home/maddux/Videos/backup/BLACK_SAILS_DISC1/BDMV/STREAM/*.m2ts")

	hash := md5.New()
	for _, file := range files {
		logger.Infof("Opening file: %s", file)
		fileStats, _ := os.Stat(file)

		bs := make([]byte, 8)
		logger.Infoln(uint64(fileStats.Size()))
		binary.LittleEndian.PutUint64(bs, uint64(fileStats.Size()))
		hash.Write(bs)
		
		// scanner := bufio.NewScanner(temp)
		// for scanner.Scan() {
		// 	line := scanner.Text()
		// 	if (!strings.HasPrefix(line, "HSH")) {
		// 		continue
		// 	}
		//
		// 	logger.Infoln(line)
		//
		// 	size, _ := strconv.ParseUint(strings.TrimSpace(strings.Split(line, ",")[3]), 10, 64)
		// 	logger.Infoln(size)
		// 	bs := make([]byte, 8)
		// 	binary.LittleEndian.PutUint64(bs, size)
		//
		// 	hash.Write(bs)
		// }

		// temp.Close()
	}

	return strings.ToUpper(fmt.Sprintf("%x", hash.Sum(nil)))
}
