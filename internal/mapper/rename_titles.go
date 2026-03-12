package mapper

import (
	"os"
	"path/filepath"
)

func RenameTitles(sourceDir string, destDir string, mappings map[string]string) []error {
    errors := []error {};
    for makemkvFileName, outputFileName := range mappings {
        srcFilePath := filepath.Join(sourceDir, makemkvFileName)
        destFilePath := filepath.Join(sourceDir, outputFileName)
        
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
