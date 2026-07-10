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

func (c *Client) RipTitle(
	ctx context.Context,
	discRoot string,
	outputDir string,
	titleId lines.TitleId,
	onLine LineSink,
) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultChan := c.runCmd(cancelCtx, "mkv", discRoot, fmt.Sprint(titleId), outputDir)

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

func (c *Client) BackupDisc(
	ctx context.Context,
	discRoot string,
	outputDir string,
	onLine LineSink,
) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultChan := c.runCmd(cancelCtx, "backup", "--decrypt", discRoot, outputDir)

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

type DiscInfo struct {
	Label  string
	Titles []Title
}

// TODO: Bring this struct's name in line with DiscInfo
type Title struct {
	SourceFilename string
	OutputFilename string
	Signature      signature.SegmentSignature
	OutputFileSize uint64
	TitleId        lines.TitleId
}

func (c *Client) ReadDisc(ctx context.Context, discRoot string) (DiscInfo, error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultChan := c.runCmd(cancelCtx, "info", discRoot)

	disc := DiscInfo{}
	titleMap := map[lines.TitleId]*Title{}
	for result := range resultChan {
		if result.Error != nil {
			return DiscInfo{}, result.Error
		}
		if result.Line == nil {
			continue
		}

		switch line := result.Line.(type) {
		case lines.TitleInfo:
			if _, exists := titleMap[line.TitleId]; !exists {
				titleMap[line.TitleId] = &Title{TitleId: line.TitleId}
			}
			title := titleMap[line.TitleId]
			if err := c.composeTitle(title, line); err != nil {
				return DiscInfo{}, err
			}
		case lines.DiscInfo:
			if line.Id == lines.DiscInfoVolumeName {
				disc.Label = line.Value
			}
		}
	}

	disc.Titles = make([]Title, 0, len(titleMap))
	for _, title := range titleMap {
		disc.Titles = append(disc.Titles, *title)
	}

	return disc, nil
}

func (c *Client) composeTitle(title *Title, line lines.TitleInfo) error {
	switch line.AttributeId {
	case lines.TitleInfoCodeSourceFileName:
		title.SourceFilename = line.Value
	case lines.TitleInfoCodeOutputFileName:
		title.OutputFilename = line.Value
	case lines.TitleInfoCodeSegmentsMap:
		signature, err := signature.NormalizeSegments(line.Value)
		if err != nil {
			return err
		}
		title.Signature = signature
	case lines.TitleInfoCodeSize:
		size, err := strconv.ParseUint(line.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing title size %q: %w", line.Value, err)
		} else {
			title.OutputFileSize = size
		}
	}

	return nil
}

func (c *Client) ScanDrives(ctx context.Context) ([]lines.DriveScan, error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultChan := c.runCmd(cancelCtx, "info", "disc:9999")

	var drives []lines.DriveScan
	for result := range resultChan {
		if result.Error != nil {
			return nil, result.Error
		} else if result.Line != nil {
			driveScan, ok := result.Line.(lines.DriveScan)
			if !ok {
				continue
			}
			drives = append(drives, driveScan)
		}
	}
	return drives, nil
}
