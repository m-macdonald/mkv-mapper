package makemkv

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)


var titleParser = regexp.MustCompile(`(.*):\d+,(\d+),\d+,"(.*)"`)

type cmdResult struct {
    Line    lines.ParsedLine
    Error   error
}

func runCmd(ctx context.Context, makeMkvPath string, arg ...string) <-chan cmdResult {
    lineProcessor := lines.NewLineProcessor()
	// TODO: Fix magic number
    resultChan := make(chan cmdResult, 32)

    go func() {
        defer close(resultChan)
        cmd := exec.CommandContext(ctx, makeMkvPath, arg...)
		cmd.Stderr = os.Stderr
        stdOutPipe, err := cmd.StdoutPipe()
        if err != nil {
            sugaredError := fmt.Errorf("Failed to establish a StdoutPipe for makemkv: %w", err)
            resultChan <- cmdResult{Error: sugaredError}

            return
        }
        if err = cmd.Start(); err != nil {
            resultChan <- cmdResult{Error: err}

            return
        }

        scanner := bufio.NewScanner(stdOutPipe)
        for scanner.Scan() {
            parsedLine, err := lineProcessor.ProcessLine(scanner.Text())

			result := cmdResult{Line: parsedLine, Error: err}
			
			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
        }
		if err := scanner.Err(); err != nil {
			resultChan <- cmdResult{Error: err}
			return
		}
        if err := cmd.Wait(); err != nil {
            resultChan <- cmdResult{Error: err}
        }
    }()

    return resultChan
}

func getDiscInfo(logger *zap.SugaredLogger, makeMkvPath string) ([]lines.ParsedLine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := runCmd(ctx, makeMkvPath, "--minlength=0", "--robot", "info", "/home/maddux/Videos/backup/BLACK_SAILS_DISC1/BDMV/index.bdmv")

	var parsedLines []lines.ParsedLine
	for line := range resultChan {
		if line.Error != nil {
			cancel()
			return nil, line.Error
		}

		parsedLines = append(parsedLines, line.Line)
	}

	return parsedLines, nil
}

func GetDiscHash(logger *zap.SugaredLogger, makeMkvPath string) (string, error) {
	logger.Debugln("Starting disc hash")


	discLines, err := getDiscInfo(logger, makeMkvPath)
	if err != nil {
		return "", err
	}

	hash := md5.New()

	for _, line := range discLines {
		logger.Infoln(line)

		titleInfo, ok := line.(lines.TitleInfo)
		if !ok {
			continue
		}

		if titleInfo.AttributeId != lines.TitleInfoCodeSize {
			continue
		}
		
		// logger.Infoln(titleInfo.Raw())
		// logger.Infoln(titleInfo.Value)

		size, err := strconv.ParseUint(titleInfo.Value, 10, 64)
		if err != nil {
			return "", err
		}

		logger.Debugf("Determined title size of: %d\n", size)
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, size)
		hash.Write(bs)
	}

	hashString := strings.ToUpper(fmt.Sprintf("%x", hash.Sum(nil)))
	
	logger.Debugf("Determined hash string: %s\n", hashString)
	return hashString, nil
}

func RipDisc(logger *zap.SugaredLogger, makeMkvPath string, opticalDriveNum int, destDir string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
    resultChan := runCmd(ctx, makeMkvPath, "mkv", fmt.Sprintf("disc:%d", opticalDriveNum), "all", destDir, "--robot")

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

func ReadTitles(makeMkvPath string) (map[string]string, error) {
    // resultChan := runCmd(makeMkvPath, "info", fmt.Sprintf("disc:%d", opticalDriveNum), "--robot")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
    resultChan := runCmd(ctx, makeMkvPath, "info", "~/Videos/backup/BLACK_SAILS_DISC1/BDMV/index.bdmv", "--robot")
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
    }

	logger.Infoln("Title read complete")

    return titles, nil
}
