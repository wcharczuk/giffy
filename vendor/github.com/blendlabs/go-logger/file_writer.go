package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	exception "github.com/blendlabs/go-exception"
)

const (
	isArchiveFileRegexpFormat           = `%s\.([0-9]+)?`
	isCompressedArchiveFileRegexpFormat = `%s\.([0-9]+)?\.gz`

	// Kilobyte represents the bytes in a kilobyte.
	Kilobyte int64 = 1 << 10
	// Megabyte represents the bytes in a megabyte.
	Megabyte int64 = Kilobyte << 10
	// Gigabyte represents the bytes in a gigabyte.
	Gigabyte int64 = Megabyte << 10
)

const (
	//FileWriterUnlimitedSize is a preset for the size of the files that can be written to be unlimited.
	FileWriterUnlimitedSize int64 = 0

	// FileWriterDefaultFileSize is the default file size (50mb).
	FileWriterDefaultFileSize int64 = 50 * Megabyte

	// FileWriterUnlimitedArchiveFiles is a preset for the number of archive files to be kept to be unlimited.
	FileWriterUnlimitedArchiveFiles int64 = 0

	// FileWriterDefaultMaxArchiveFiles is the default number of archive files (10).
	FileWriterDefaultMaxArchiveFiles int64 = 10
)

// NewFileWriterFromEnvironmentVars creates a new FileWriter from the given environment variable names.`
func NewFileWriterFromEnvironmentVars(pathVar, shouldCompressVar, maxSizeVar, maxArchiveVar string) (*FileWriter, error) {
	filePath := os.Getenv(pathVar)
	if len(filePath) == 0 {
		return nil, fmt.Errorf("Environment Variable `%s` required", pathVar)
	}

	shouldCompress := envFlagIsSet(shouldCompressVar, false)
	maxFileSize := File.ParseSize(os.Getenv(maxSizeVar), FileWriterDefaultFileSize)
	maxArchive := envFlagInt64(maxArchiveVar, FileWriterDefaultMaxArchiveFiles)
	return NewFileWriter(filePath, shouldCompress, maxFileSize, maxArchive)
}

// NewFileWriterWithDefaults returns a new file writer with defaults.
func NewFileWriterWithDefaults(filePath string) (*FileWriter, error) {
	return NewFileWriter(filePath, true, FileWriterDefaultFileSize, FileWriterDefaultMaxArchiveFiles)
}

// NewFileWriter creates a new file writer.
func NewFileWriter(filePath string, shouldCompressArchivedFiles bool, fileMaxSizeBytes, fileMaxArchiveCount int64) (*FileWriter, error) {
	file, err := File.CreateOrOpen(filePath)
	if err != nil {
		return nil, err
	}

	var regex *regexp.Regexp
	if shouldCompressArchivedFiles {
		regex, err = createIsCompressedArchiveFileRegexp(filePath)

	} else {
		regex, err = createIsArchivedFileRegexp(filePath)
	}
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		filePath:                    filePath,
		file:                        file,
		syncRoot:                    &sync.Mutex{},
		isArchiveFileRegexp:         regex,
		shouldCompressArchivedFiles: shouldCompressArchivedFiles,
		fileMaxSizeBytes:            fileMaxSizeBytes,
		fileMaxArchiveCount:         fileMaxArchiveCount,
	}, nil
}

// FileWriter implements the file rotation settings from supervisor.
type FileWriter struct {
	filePath string
	file     *os.File

	syncRoot                    *sync.Mutex
	shouldCompressArchivedFiles bool

	fileMaxSizeBytes    int64
	fileMaxArchiveCount int64

	isArchiveFileRegexp *regexp.Regexp
}

// Write writes to the file.
func (fw *FileWriter) Write(buffer []byte) (int, error) {
	fw.syncRoot.Lock()
	defer fw.syncRoot.Unlock()

	if fw.fileMaxSizeBytes > 0 {
		stat, err := fw.file.Stat()
		if err != nil {
			return 0, exception.New(err)
		}

		if stat.Size() > fw.fileMaxSizeBytes {
			err = fw.rotateFile()
			if err != nil {
				return 0, exception.New(err)
			}
		}
	}

	written, err := fw.file.Write(buffer)
	return written, exception.Wrap(err)
}

// Close closes the stream.
func (fw *FileWriter) Close() error {
	if fw.file != nil {
		err := fw.file.Close()
		fw.file = nil
		return err
	}
	return nil
}

func (fw *FileWriter) makeArchiveFilePath(filePath string, index int64) string {
	return fmt.Sprintf("%s.%d", filePath, index)
}

func (fw *FileWriter) makeCompressedArchiveFilePath(filePath string, index int64) string {
	return fmt.Sprintf("%s.%d.gz", filePath, index)
}

func (fw *FileWriter) makeTempArchiveFilePath(filePath string, index int64) string {
	return fmt.Sprintf("%s.%d.tmp", filePath, index)
}

func (fw *FileWriter) makeTempCompressedArchiveFilePath(filePath string, index int64) string {
	return fmt.Sprintf("%s.%d.gz.tmp", filePath, index)
}

func (fw *FileWriter) compressFile(inFilePath, outFilePath string) error {
	inFile, err := os.Open(inFilePath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzw := gzip.NewWriter(outFile)
	defer gzw.Close()

	_, err = io.Copy(gzw, inFile)
	if err != nil {
		return err
	}
	return gzw.Flush()
}

func (fw *FileWriter) extractArchivedFileIndex(filePath string) (int64, error) {
	filePathBase := filepath.Base(filePath)
	values := fw.isArchiveFileRegexp.FindStringSubmatch(filePathBase)
	if len(values) > 1 {
		value, err := strconv.ParseInt(values[1], 10, 32)
		return value, exception.Wrap(err)
	}
	return 0, exception.Newf("Cannot extract file index from `%s`", filePathBase)
}

func (fw *FileWriter) isArchivedFile(filePath string) bool {
	return fw.isArchiveFileRegexp.MatchString(filePath)
}

func (fw *FileWriter) getArchivedFilePaths() ([]string, error) {
	return File.List(filepath.Dir(fw.filePath), fw.isArchiveFileRegexp)
}

func (fw *FileWriter) getMaxArchivedFileIndex(paths []string) (int64, error) {
	var err error
	var index int64
	max := int64(-1 << 63)
	for _, path := range paths {
		index, err = fw.extractArchivedFileIndex(path)
		if err != nil {
			return max, err
		}
		if index > max {
			max = index
		}
	}
	return max, err
}

func (fw *FileWriter) shiftArchivedFiles(paths []string) error {
	var index int64
	var err error

	intermediatePaths := make(map[string]string)
	var tempPath, finalPath string

	for _, path := range paths {
		index, err = fw.extractArchivedFileIndex(path)
		if err != nil {
			return err
		}
		if fw.shouldCompressArchivedFiles {
			tempPath = fw.makeTempCompressedArchiveFilePath(fw.filePath, index+1)
			finalPath = fw.makeCompressedArchiveFilePath(fw.filePath, index+1)
		} else {
			tempPath = fw.makeTempArchiveFilePath(fw.filePath, index+1)
			finalPath = fw.makeArchiveFilePath(fw.filePath, index+1)
		}

		if fw.fileMaxArchiveCount > 0 {
			if index+1 <= fw.fileMaxArchiveCount {
				err = os.Rename(path, tempPath)
				intermediatePaths[tempPath] = finalPath
			} else {
				err = os.Remove(path)
			}
		} else {
			err = os.Rename(path, tempPath)
			intermediatePaths[tempPath] = finalPath
		}
		if err != nil {
			return err
		}
	}

	for from, to := range intermediatePaths {
		err = os.Rename(from, to)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fw *FileWriter) rotateFile() error {
	var err error

	paths, err := fw.getArchivedFilePaths()
	if err != nil {
		return err
	}

	err = fw.shiftArchivedFiles(paths)
	if err != nil {
		return err
	}

	err = fw.file.Close()
	if err != nil {
		return err
	}

	if fw.shouldCompressArchivedFiles {
		err = fw.compressFile(fw.filePath, fw.makeCompressedArchiveFilePath(fw.filePath, 1))
		if err != nil {
			return err
		}
		err = os.Remove(fw.filePath)
		if err != nil {
			return err
		}
	} else {
		err = os.Rename(fw.filePath, fw.makeArchiveFilePath(fw.filePath, 1))
	}

	file, err := os.Create(fw.filePath)
	if err != nil {
		return err
	}
	fw.file = file
	return nil
}

func createIsArchivedFileRegexp(filePath string) (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf(isArchiveFileRegexpFormat, filepath.Base(filePath)))
}

func createIsCompressedArchiveFileRegexp(filePath string) (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf(isCompressedArchiveFileRegexpFormat, filepath.Base(filePath)))
}
