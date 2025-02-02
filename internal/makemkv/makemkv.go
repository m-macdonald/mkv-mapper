package makemkv

import (
    "bufio"
    "fmt"
    "os/exec"
    "strings"
)

func splitLine(line string) []string {
    return strings.Split(line[6:], ",")
}

func RipDisc(makeMkvPath string, opticalDriveNum int, destDir string) error {
    cmd := exec.Command(makeMkvPath, "mkv", fmt.Sprintf("disc:%d", opticalDriveNum), "all", destDir, "--robot")
    _, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }
    cmd.Start()

    // TODO: Display current ripping progress

    if err = cmd.Wait(); err != nil {
        return err
    }
    
    return nil
}

func ReadTitles(makeMkvPath string, opticalDriveNum int) (map[string]string, error) {
    cmd := exec.Command(makeMkvPath, "info", fmt.Sprint("disc:%d", opticalDriveNum), "--robot")
    stdOutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err) 
    }
    cmd.Start()
    titles := make(map[string]string)
    
    scanner := bufio.NewScanner(stdOutPipe)
    for scanner.Scan() {
        line := scanner.Text()
        if (!strings.HasPrefix(line, "TINFO")) {
            continue
        }
        params := splitLine(line)
        //Might need to consider using the track number for uniqueness just while parsing these
        // 16 is the mpls name
        // 27 is the name of that makemkv will give the file
        // TODO: Clean this up
        if (params[1] == "16") {
            mplsName := params[3] 
            for params[1] != "27" && scanner.Scan() {
                line = scanner.Text()
                params = splitLine(line)
            }
            if scanner.Err() != nil {
                return nil, fmt.Errorf("%w", err)
            }

            outputName := params[3]
            
            titles[outputName] = mplsName 
        }
    }
    // TODO: If the makemkv license has expired this fails quietly. Need to give more info in the terminal in that situation
    if scanner.Err() != nil {
        return nil, fmt.Errorf("Error while scanning makemkvcon stdout: %w", err)
    }

    if err = cmd.Wait(); err != nil {
        return nil, fmt.Errorf("Error ocurred while waiting for makemkv to finish processing the disc: %w", err)
    }

    return titles, nil
}
