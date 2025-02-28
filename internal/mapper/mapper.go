package mapper

import (
	"m-macdonald/mkv-mapper/internal/discdb"
	"os"
	"path"
)

func Map(fileDir string, mappings map[string]discdb.TitleSummary) []error {
    errors := []error {};
    for fileName, titleSummary := range mappings {
        srcFilePath := path.Join(fileDir, fileName)
        destFilePath := path.Join(fileDir, titleSummary.Name)
        
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
