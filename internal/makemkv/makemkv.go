package makemkv

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

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
	discRoot string,
	outputDir string,
	onLine LineSink,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := c.runCmd(ctx, "mkv", discRoot, "all", outputDir)

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
	SourceFilename   string
	OutputFilename   string
	SegmentSignature signature.SegmentSignature
	OutputFileSize   uint64
	TitleId          int
}

func (c *Client) ReadTitles(discRoot string) (map[signature.SegmentSignature]Title, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := c.runCmd(ctx, "info", discRoot)
	titles := make(map[int]Title)

	for result := range resultChan {
		if result.Error != nil {
			return nil, result.Error
		} else if result.Line != nil {
			parsedLine := result.Line

			titleInfo, ok := parsedLine.(lines.TitleInfo)
			if !ok {
				continue
			}

			title, ok := titles[titleInfo.TitleId]
			if !ok {
				title = Title{
					TitleId: titleInfo.TitleId,
				}
			}

			switch titleInfo.AttributeId {
			case lines.TitleInfoCodeSourceFileName:
				title.SourceFilename = titleInfo.Value
			case lines.TitleInfoCodeOutputFileName:
				title.OutputFilename = titleInfo.Value
			case lines.TitleInfoCodeSegmentsMap:
				segmentSignature, err := signature.NormalizeSegments(titleInfo.Value)
				if err != nil ||
					segmentSignature == "" {
					// TODO: continue if we can't parse?
				} else {
					title.SegmentSignature = segmentSignature
				}
			case lines.TitleInfoCodeSize:
				size, err := strconv.ParseUint(titleInfo.Value, 10, 64)
				if err != nil {
				} else {
					title.OutputFileSize = size
				}
			}
			titles[titleInfo.TitleId] = title
		}
	}

	t := make(map[signature.SegmentSignature]Title)
	for _, title := range titles {
		t[title.SegmentSignature] = title
	}

	return t, nil
}
