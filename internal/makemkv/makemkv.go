package makemkv

import (
	"bufio"
	"fmt"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"os/exec"
	"regexp"

	"go.uber.org/zap"
)


var titleParser = regexp.MustCompile(`(.*):\d+,(\d+),\d+,"(.*)"`)

type cmdResult struct {
    Line    lines.ParsedLine
    Error   error
    Done    bool
}

func runCmd(makeMkvPath string, arg ...string) <-chan cmdResult {
    lineProcessor := lines.LineProcessor {}
    resultChan := make(chan cmdResult)

    go func() {
        defer close(resultChan)
        cmd := exec.Command(makeMkvPath, arg...)
        stdOutPipe, err := cmd.StdoutPipe()
        if err != nil {
            sugaredError := fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err)
            resultChan <- cmdResult{Error: sugaredError, Done: true}

            return
        }
        if err = cmd.Start(); err != nil {
            resultChan <- cmdResult{Error: err, Done: true}

            return
        }

        scanner := bufio.NewScanner(stdOutPipe)
        for scanner.Scan() {
            parsedLine, err := lineProcessor.ProcessLine(scanner.Text())
            if err != nil {
                resultChan <- cmdResult{Error: err, Done: false}
            } else {
                resultChan <- cmdResult{Line: parsedLine, Error: nil, Done: false}
            }
        }

        if err := cmd.Wait(); err != nil {
            resultChan <- cmdResult{Error: err, Done: true}
        } else {
            resultChan <- cmdResult{Done: true}
        }
    }()

    return resultChan
}

func RipDisc(logger *zap.SugaredLogger, makeMkvPath string, opticalDriveNum int, destDir string) error {
    resultChan := runCmd(makeMkvPath, "mkv", fmt.Sprintf("disc:%d", opticalDriveNum), "all", destDir, "--robot")

    for result := range resultChan {
        if result.Error != nil {
            logger.Error(result.Error)
        } else if result.Line != nil {
            // Do nothing with it yet
        }
    }





    // cmd := exec.Command(makeMkvPath, "mkv", fmt.Sprintf("disc:%d", opticalDriveNum), "all", destDir, "--robot")
    // stdOutPipe, err := cmd.StdoutPipe()
    // if err != nil {
    //     return fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err)
    // }
    // if err = cmd.Start(); err != nil {
    //     return fmt.Errorf("Failed to start ", err)
    // }
    //
    // scanner := bufio.NewScanner(stdOutPipe)
    // for scanner.Scan() {
    //     nextLine(logger, scanner)
    // }
    // 
    // if err = cmd.Wait(); err != nil {
    //     return err
    // }
    
    return nil
}

func ReadTitles(logger *zap.SugaredLogger, makeMkvPath string, opticalDriveNum int) (map[string]string, error) {
    resultChan := runCmd(makeMkvPath, "info", fmt.Sprintf("disc:%d", opticalDriveNum), "--robot")
    titles := make(map[string]string)

    // var mplsName string

    for result := range resultChan {
        if result.Error != nil {
            logger.Error(result.Error)
        } else if result.Line != nil {
            parsedLine := result.Line
            logger.Debugln(parsedLine)
			// TODO: Fix this
            // switch t := parsedLine.(type) {
            // case lines.TitleInfo:
            //     if t.Code == 16 {
            //         mplsName = t.Value
            //     } else if t.Code == 27 {
            //         titles[mplsName] = t.Value
            //     }
            // }
        }

        if (result.Done) {
            logger.Debugln("Title read complete")
            break
        }
    }

    return titles, nil
}

// func ReadTitles(logger *zap.SugaredLogger, makeMkvPath string, opticalDriveNum int) (map[string]string, error) {
//     cmd := exec.Command(makeMkvPath, "info", fmt.Sprintf("disc:%d", opticalDriveNum), "--robot")
//     stdOutPipe, err := cmd.StdoutPipe()
//     if err != nil {
//         return nil, fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err) 
//     }
//     
//     if err = cmd.Start(); err != nil {
//         logger.Debugln("%v", cmd.Args)
//         return nil, fmt.Errorf("Failed to start reading titles from disc: %w", err)
//     }
//     titles := make(map[string]string)
//
//     lineProcessor := lines.LineProcessor {}
//     
//     scanner := bufio.NewScanner(stdOutPipe)
//     for scanner.Scan() {
//         // TODO: Need to account for the situation where the disc failed to read
//         // MSG:5010,0,0,"Failed to open disc"
//         line := nextLine(logger, scanner)
//         if (!strings.HasPrefix(line, "TINFO")) {
//             continue
//         }
//
//         parsedLine, err := lineProcessor.ProcessLine(line)
//         if err != nil {
//             // error handling
//         }
//         //Might need to consider using the track number for uniqueness just while parsing these
//         // 16 is the mpls name
//         // 27 is the name of that makemkv will give the file
//         // TODO: Clean this up
//         // TODO: Check that matches is not nil
//         switch t := parsedLine.(type) {
//         case lines.TitleInfo: 
//             if t.Code == 16 {
//                 mplsName := t.Value
//                 for t.Code != 27 && scanner.Scan() {
//                     line = nextLine(logger, scanner)
//                     fileNameLine, err := lineProcessor.ProcessLine(line)
//                     switch t := fileNameLine.(type) {
//                     case lines.TitleInfo:
//                         if t.Code == 27 {
//                             outputName := t.Value
//                         }
//                     }
//                 }
//             }
//         }
//         // if (matches[2] == "16") {
//         //     mplsName := matches[3]
//         //     for matches == nil || matches[2] != "27" && scanner.Scan() {
//         //         line = nextLine(logger, scanner)
//         //         matches = titleParser.FindStringSubmatch(line)
//         //     }
//         //     if scanner.Err() != nil {
//         //         return nil, fmt.Errorf("%w", err)
//         //     }
//         //
//         //     outputName := matches[3]
//             
//         titles[mplsName] = outputName 
//     }
//     // TODO: If the makemkv license has expired this fails quietly. Need to give more info in the terminal in that situation
//     if scanner.Err() != nil {
//         return nil, fmt.Errorf("Error while scanning makemkvcon stdout: %w", err)
//     }
//
//     if err = cmd.Wait(); err != nil {
//         return nil, fmt.Errorf("Error occurred while waiting for makemkv to finish processing the disc: %w", err)
//     }
//
//     return titles, nil
// }
//
// func nextLine(logger *zap.SugaredLogger, scanner *bufio.Scanner) string {
//     line := scanner.Text()
//     logger.Debugln(line)
//     
//     return line
// }
