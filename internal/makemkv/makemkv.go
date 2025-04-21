package makemkv

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)


var titleParser = regexp.MustCompile(`(.*):\d+,(\d+),\d+,"(.*)"`)

func RipDisc(logger *zap.SugaredLogger, makeMkvPath string, opticalDriveNum int, destDir string) error {
    cmd := exec.Command(makeMkvPath, "mkv", fmt.Sprintf("disc:%d", opticalDriveNum), "all", destDir, "--robot")
    stdOutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err)
    }
    if err = cmd.Start(); err != nil {
        return fmt.Errorf("Failed to start ", err)
    }

    scanner := bufio.NewScanner(stdOutPipe)
    for scanner.Scan() {
        nextLine(logger, scanner)
    }
    
    if err = cmd.Wait(); err != nil {
        return err
    }
    
    return nil
}

func ReadTitles(logger *zap.SugaredLogger,makeMkvPath string, opticalDriveNum int) (map[string]string, error) {
    cmd := exec.Command(makeMkvPath, "info", fmt.Sprintf("disc:%d", opticalDriveNum), "--robot")
    stdOutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err) 
    }
    
    if err = cmd.Start(); err != nil {
        return nil, fmt.Errorf("Failed to start reading titles from disc: %w", err)
    }
    titles := make(map[string]string)
    
    scanner := bufio.NewScanner(stdOutPipe)
    for scanner.Scan() {
        // TODO: Need to account for the situation where the disc failed to read
        // MSG:5010,0,0,"Failed to open disc"
        // When debug is enabled should probably write the makemkv output
        line := nextLine(logger, scanner)
        if (!strings.HasPrefix(line, "TINFO")) {
            continue
        }

        matches := titleParser.FindStringSubmatch(line)
        //Might need to consider using the track number for uniqueness just while parsing these
        // 16 is the mpls name
        // 27 is the name of that makemkv will give the file
        // TODO: Clean this up
        // TODO: Check that matches is not nil
        if (matches[2] == "16") {
            mplsName := matches[3]
            for matches == nil || matches[2] != "27" && scanner.Scan() {
                line = nextLine(logger, scanner)
                matches = titleParser.FindStringSubmatch(line)
            }
            if scanner.Err() != nil {
                return nil, fmt.Errorf("%w", err)
            }

            outputName := matches[3]
            
            titles[mplsName] = outputName 
        }
    }
    // TODO: If the makemkv license has expired this fails quietly. Need to give more info in the terminal in that situation
    if scanner.Err() != nil {
        return nil, fmt.Errorf("Error while scanning makemkvcon stdout: %w", err)
    }

    if err = cmd.Wait(); err != nil {
        return nil, fmt.Errorf("Error occurred while waiting for makemkv to finish processing the disc: %w", err)
    }

    return titles, nil
}

func nextLine(logger *zap.SugaredLogger, scanner *bufio.Scanner) string {
    line := scanner.Text()
    logger.Debugln(line)
    
    return line
}
