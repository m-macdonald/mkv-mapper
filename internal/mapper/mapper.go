package mapper

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"m-macdonald/mkv-mapper/internal/common"
	"m-macdonald/mkv-mapper/internal/discdb"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func Map(fileDir string, mappings map[string]discdb.TitleSummary) []error {
    errors := []error {};
    for fileName, titleSummary := range mappings {
        srcFilePath := filepath.Join(fileDir, fileName)
        destFilePath := filepath.Join(fileDir, titleSummary.Name)
        
        err := os.Rename(srcFilePath, destFilePath)

        if err != nil {
            errors = append(errors, err)
        }
    } 

    if len(errors) > 0 {
        return errors
    }

    return nil
}
