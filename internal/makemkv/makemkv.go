package makemkv

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/signature"

	"go.uber.org/zap"
)

type Client struct {
	makeMkvPath string
	logger      *zap.SugaredLogger
}

func NewClient(makeMkvPath string, logger *zap.SugaredLogger) *Client {
	return &Client{
		makeMkvPath: makeMkvPath,
		logger:      logger,
	}
}

type cmdResult struct {
	Line  lines.ParsedLine
	Error error
}

func (c *Client) runCmd(ctx context.Context, arg ...string) <-chan cmdResult {
	lineProcessor := lines.NewLineProcessor()
	// TODO: Fix magic number
	resultChan := make(chan cmdResult, 32)

	go func() {
		defer close(resultChan)
		cmd := exec.CommandContext(ctx, c.makeMkvPath, arg...)
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

func (c *Client) getDiscInfo() ([]lines.ParsedLine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := c.runCmd(ctx, "--minlength=0", "--robot", "info", "/home/maddux/Videos/backup/BLACK_SAILS_DISC1/BDMV/index.bdmv")

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

func (c *Client) RipDisc() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := c.runCmd(ctx, "mkv", "all", "destDir", "--robot")

	for result := range resultChan {
		if result.Error != nil {
			c.logger.Error(result.Error)
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

type makeMkvTitle struct {
	SourceFileName   string
	OutputFileName   string
	SegmentSignature signature.SegmentSignature
}

func (c *Client) ReadTitles(discRoot string) (map[signature.SegmentSignature]string, error) {
	// resultChan := runCmd(makeMkvPath, "info", fmt.Sprintf("disc:%d", opticalDriveNum), "--robot")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := c.runCmd(ctx, "info", "/home/maddux/Videos/backup/BLACK_SAILS_DISC1/BDMV/index.bdmv", "--robot")
	titles := make(map[uint]makeMkvTitle)

	for result := range resultChan {
		if result.Error != nil {
			return nil, result.Error
		} else if result.Line != nil {
			parsedLine := result.Line

			titleInfo, ok := parsedLine.(lines.TitleInfo)
			if !ok {
				continue
			}

			titleFileNames, ok := titles[titleInfo.TitleId]
			if !ok {
				titleFileNames = makeMkvTitle{}
			}

			switch titleInfo.AttributeId {
			case lines.TitleInfoCodeSourceFileName:
				titleFileNames.SourceFileName = titleInfo.Value
			case lines.TitleInfoCodeOutputFileName:
				titleFileNames.OutputFileName = titleInfo.Value
			case lines.TitleInfoCodeSegmentsMap:
				segmentSignature, err := signature.NormalizeSegments(titleInfo.Value)
				if err != nil ||
					segmentSignature == "" {
					// continue if we can't parse?
				} else {
					titleFileNames.SegmentSignature = segmentSignature
				}
			}
			titles[titleInfo.TitleId] = titleFileNames
		}
	}

	t := make(map[signature.SegmentSignature]string)
	for _, v := range titles {
		t[v.SegmentSignature] = v.OutputFileName
	}

	return t, nil
}
