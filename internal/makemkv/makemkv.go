package makemkv

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"

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

func (c *Client) runCmd(ctx context.Context, args ...string) <-chan cmdResult {
	lineProcessor := lines.NewLineProcessor()
	// TODO: Fix magic number
	resultChan := make(chan cmdResult, 32)

	fullArgs := append([]string{}, "--robot", "--progress=-stdout")
	fullArgs = append(fullArgs, args...)

	go func() {
		defer close(resultChan)
		cmd := exec.CommandContext(ctx, c.makeMkvPath, fullArgs...)
		cmd.Stdin = nil
		cmd.Stderr = os.Stderr
		stdOutPipe, err := cmd.StdoutPipe()
		if err != nil {
			sugaredError := fmt.Errorf("failed to establish a StdoutPipe for makemkv: %w", err)
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

type LineSink func(lines.ParsedLine)

func (c *Client) RipDisc(
	ctx context.Context,
	discRoot string,
	outputDir string,
	onLine LineSink,
) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultChan := c.runCmd(cancelCtx, "mkv", discRoot, "all", outputDir)

	for result := range resultChan {
		if result.Error != nil {
			cancel()

			return result.Error
		}

		if result.Line != nil {
			onLine(result.Line)
		}
	}

	return nil
}

type Title struct {
	SourceFilename string
	OutputFilename string
	Segments       string
	OutputFileSize uint64
	TitleId        int
}

func (c *Client) ReadTitles(ctx context.Context, discRoot string) ([]Title, error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultChan := c.runCmd(cancelCtx, "info", discRoot)
	titles := []Title{}

	for result := range resultChan {
		if result.Error != nil {
			return nil, result.Error
		} else if result.Line != nil {
			parsedLine := result.Line

			titleInfo, ok := parsedLine.(lines.TitleInfo)
			if !ok {
				continue
			}

			title := Title{
				TitleId: titleInfo.TitleId,
			}

			switch titleInfo.AttributeId {
			case lines.TitleInfoCodeSourceFileName:
				title.SourceFilename = titleInfo.Value
			case lines.TitleInfoCodeOutputFileName:
				title.OutputFilename = titleInfo.Value
			case lines.TitleInfoCodeSegmentsMap:
				title.Segments = titleInfo.Value
			case lines.TitleInfoCodeSize:
				size, err := strconv.ParseUint(titleInfo.Value, 10, 64)
				if err != nil {
				} else {
					title.OutputFileSize = size
				}
			}
			titles = append(titles, title)
		}
	}

	return titles, nil
}
