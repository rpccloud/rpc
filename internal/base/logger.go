package base

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Logger ...
type Logger struct {
	isLogToScreen bool
	file          *os.File
	mu            sync.Mutex
}

// NewLogger ...
func NewLogger(isLogToScreen bool, outFile string) (*Logger, *Error) {
	file, e := func() (*os.File, error) {
		if outFile == "" {
			return nil, nil
		}

		// make sure the dir of outFile is exist
		dirName := filepath.Dir(outFile)

		if e := os.Mkdir(dirName, os.ModeDir|0755); e != nil {
			if !os.IsExist(e) {
				return nil, e
			}

			if info, e := os.Stat(dirName); e != nil || !info.IsDir() {
				return nil, fmt.Errorf("path %s is not a directory", dirName)
			}
		}

		return os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}()

	if e != nil {
		return &Logger{
			isLogToScreen: isLogToScreen,
			file:          nil,
		}, ErrLogOpenFile.AddDebug(e.Error())
	}

	return &Logger{
		isLogToScreen: isLogToScreen,
		file:          file,
	}, nil
}

// Log ...
func (p *Logger) Log(str string) {
	if p.file != nil {
		_, _ = p.file.WriteString(ConcatString(str, "\n"))
	}

	if p.isLogToScreen {
		_, _ = os.Stdout.WriteString(ConcatString(str, "\n"))
	}
}

// Close ...
func (p *Logger) Close() *Error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.file != nil {
		if e := p.file.Close(); e != nil {
			return ErrLogCloseFile.AddDebug(e.Error())
		}

		p.file = nil
		return nil
	}

	return nil
}
