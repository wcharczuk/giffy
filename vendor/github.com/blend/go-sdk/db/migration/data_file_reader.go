package migration

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/util"
	"github.com/lib/pq"
)

const (
	runeTab          = rune('\t')
	runeNewline      = rune('\n')
	byteNewLine      = byte('\n')
	byteTab          = byte('\t')
	null             = `\N`
	regexCopyExtract = `COPY (.*)? \((.*)?\)`
)

// ReadDataFile returns a new DataFileReader
func ReadDataFile(filePath string) *DataFileReader {
	return &DataFileReader{
		path:          filePath,
		copyExtractor: regexp.MustCompile(regexCopyExtract),
	}
}

// DataFileReader reads a postgres dump.
type DataFileReader struct {
	parent Migration
	label  string
	path   string
	logger *Logger

	copyExtractor *regexp.Regexp
}

// Label returns the label for the data file reader.
func (dfr *DataFileReader) Label() string {
	if len(dfr.label) == 0 {
		dfr.label = fmt.Sprintf("read data file `%s`", dfr.path)
	}
	return dfr.label
}

// SetLabel sets the migration label.
func (dfr *DataFileReader) SetLabel(value string) {
	dfr.label = value
}

// WithLabel sets the migration label.
func (dfr *DataFileReader) WithLabel(value string) Migration {
	dfr.label = value
	return dfr
}

// Parent returns the parent for the data file reader.
func (dfr *DataFileReader) Parent() Migration {
	return dfr.parent
}

// SetParent sets the parent for the data file reader.
func (dfr *DataFileReader) SetParent(parent Migration) {
	dfr.parent = parent
}

// WithParent sets the parent for the data file reader.
func (dfr *DataFileReader) WithParent(parent Migration) Migration {
	dfr.parent = parent
	return dfr
}

// Logger returns the logger.
func (dfr *DataFileReader) Logger() *Logger {
	return dfr.logger
}

// SetLogger sets the logger for the data file reader.
func (dfr *DataFileReader) SetLogger(logger *Logger) {
	dfr.logger = logger
}

// WithLogger sets the logger for the data file reader.
func (dfr *DataFileReader) WithLogger(logger *Logger) Migration {
	dfr.logger = logger
	return dfr
}

// IsTransactionIsolated returns if the migration is transaction isolated or not.
func (dfr *DataFileReader) IsTransactionIsolated() bool {
	return true
}

// Test runs the data file reader and then rolls-back the txn.
func (dfr *DataFileReader) Test(c *db.Connection, optionalTx ...*sql.Tx) (err error) {
	tx, err := c.Begin()
	if err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", err)
		}
		if err == nil {
			dfr.logger.Applyf(dfr, "done")
		} else {
			dfr.logger.Error(dfr, err)
		}
		tx.Rollback()
	}()
	err = dfr.Invoke(c, tx)
	return
}

// Apply applies the data file reader.
func (dfr *DataFileReader) Apply(c *db.Connection, optionalTx ...*sql.Tx) (err error) {
	tx, err := c.Begin()
	if err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", err)
		}
		if err == nil {
			tx.Commit()
			dfr.logger.Applyf(dfr, "done")
		} else {
			tx.Rollback()
			dfr.logger.Error(dfr, err)
		}
	}()

	err = dfr.Invoke(c, tx)
	return
}

// Invoke consumes the data file and writes it to the db.
func (dfr *DataFileReader) Invoke(c *db.Connection, tx *sql.Tx) (err error) {
	var f *os.File
	if f, err = os.Open(dfr.path); err != nil {
		return
	}
	defer f.Close()

	var stmt *sql.Stmt
	var state int

	var cursor int64
	var readBuffer = make([]byte, 32)
	var readErr error
	var lineBuffer = bytes.NewBuffer([]byte{})
	var line string
	var pieces []interface{}

	for readErr == nil {
		lineBuffer.Reset()

		switch state {
		case 0:
			cursor, readErr = dfr.readLine(f, cursor, readBuffer, lineBuffer)
			if readErr != nil {
				continue
			}
			line = lineBuffer.String()

			if util.String.HasPrefixCaseInsensitive(line, "--") {
				continue
			}

			if util.String.HasPrefixCaseInsensitive(line, "set") {
				continue
			}

			if util.String.HasPrefixCaseInsensitive(line, "copy") {
				if !util.String.HasSuffixCaseInsensitive(line, "from stdin;") {
					err = fmt.Errorf("only `stdin` from clauses supported at this time, cannot continue")
					return
				}

				stmt, err = dfr.executeCopyLine(line, c, tx)
				if err != nil {
					return
				}
				state = 1
				continue
			}

			err = c.ExecInTx(line, tx)
			if err != nil {
				return
			}
		case 1:
			pieces, cursor, readErr = dfr.readTabLine(f, cursor, readBuffer, lineBuffer)
			if readErr != nil {
				continue
			}

			if len(pieces) == 0 {
				err = fmt.Errorf("empty data line, we error on this for now")
				return
			}

			if len(pieces) == 1 && util.String.HasPrefixCaseInsensitive(pieces[0].(string), `\.`) {
				err = stmt.Close()
				if err != nil {
					return
				}
				state = 0
				continue
			}

			_, err = stmt.Exec(pieces...)
			if err != nil {
				return
			}
		}
	}
	return nil
}

func (dfr *DataFileReader) executeCopyLine(line string, c *db.Connection, tx *sql.Tx) (*sql.Stmt, error) {
	pieces := dfr.extractCopyLine(line)
	if len(pieces) < 3 {
		return nil, exception.New("Invalid `COPY ...` line, cannot continue.")
	}
	tableName := pieces[1]
	columnCSV := pieces[2]
	columns := strings.Split(columnCSV, ", ")
	return tx.Prepare(pq.CopyIn(tableName, columns...))
}

// regexExtractSubMatches returns sub matches for an expr because go's regexp library is weird.
func (dfr *DataFileReader) extractCopyLine(line string) []string {
	allResults := dfr.copyExtractor.FindAllStringSubmatch(line, -1)
	results := []string{}
	for _, resultSet := range allResults {
		for _, result := range resultSet {
			results = append(results, result)
		}
	}

	return results
}

func (dfr *DataFileReader) extractDataLine(line string) []interface{} {
	var values []interface{}
	var value string
	var state int

	appendValue := func() {
		if value == `\N` {
			values = append(values, nil)
		} else {
			values = append(values, value)
		}
	}

	for _, r := range line {
		switch state {
		case 0:
			if r == runeTab {
				continue
			}
			state = 1
			value = value + string(r)
		case 1:
			if r == runeTab {
				appendValue()
				state = 0
				value = ""
				continue
			}

			value = value + string(r)
		}
	}

	if len(value) > 0 {
		appendValue()
	}

	return values
}

// readLine reads a file until a newline.
func (dfr *DataFileReader) readLine(f io.ReaderAt, cursor int64, readBuffer []byte, lineBuffer *bytes.Buffer) (int64, error) {
	// bytesRead is the return from the ReadAt function
	// it indicates how many effective bytes we read from the stream.
	var bytesRead int
	// err is our primary indicator if there was an issue with the stream
	// or if we've reached the end of the file.
	var err error
	// b is the byte we're reading at a time.
	var b byte

	// while we haven't hit an error (this includes EOF!)
	for err == nil {
		// read the stream
		bytesRead, err = f.ReadAt(readBuffer, cursor)
		// abort on error
		if err != nil && err != io.EOF { //let this continue on eof
			return cursor, err
		}

		// slurp the read buffer.
		for readBufferIndex := 0; readBufferIndex < bytesRead; readBufferIndex++ {
			// advance the cursor regardless of what we read out.
			// if we read a newline, great! we'll start the next character after the newline after.
			cursor++

			// slurp the byte out of the read buffer
			b = readBuffer[readBufferIndex]
			if b == byteNewLine {
				// we bifurcate here because we need to forward the eof
				// if we read the buffer exactly right.
				if readBufferIndex == bytesRead-1 {
					return cursor, err
				}
				// otherwise the newline may have happened
				// before the actual eof.
				return cursor, nil
			}

			// b wasnt a newline, write it to the output buffer.
			lineBuffer.WriteByte(b)
		}
	}
	// we've reached the end of the file
	// there may not have been a newline
	// return what we have
	return cursor, err
}

// readTabLine reads a file until a new line, collecting tab delimited sections into an array.
func (dfr *DataFileReader) readTabLine(f io.ReaderAt, cursor int64, readBuffer []byte, lineBuffer *bytes.Buffer) ([]interface{}, int64, error) {
	// bytesRead is the return from the ReadAt function
	// it indicates how many effective bytes we read from the stream.
	var bytesRead int
	// err is our primary indicator if there was an issue with the stream
	// or if we've reached the end of the file.
	var err error
	// b is the byte we're reading at a time.
	var b byte
	// pieces is used to collect the tab delimited components of the line.
	var pieces []interface{}
	// while we haven't hit an error (this includes EOF!)
	for err == nil {
		// read the stream
		bytesRead, err = f.ReadAt(readBuffer, cursor)
		// abort on error
		if err != nil && err != io.EOF { //let this continue on eof
			return pieces, cursor, err
		}

		// slurp the read buffer.
		for readBufferIndex := 0; readBufferIndex < bytesRead; readBufferIndex++ {
			// advance the cursor regardless of what we read out.
			// if we read a newline, great! we'll start the next character after the newline after.
			cursor++

			// slurp the byte out of the read buffer
			b = readBuffer[readBufferIndex]
			if b == byteNewLine {
				// make sure to collect the remaining text in the
				// linebuffer.
				pieces = dfr.readTabLineAppendPiece(pieces, lineBuffer)

				// we bifurcate here because we need to forward the eof
				// if we read the buffer exactly right.
				if readBufferIndex == bytesRead-1 {
					return pieces, cursor, err
				}
				// otherwise the newline may have happened
				// before the actual eof.
				return pieces, cursor, nil
			}

			// if we see a tab
			// mark a tab section
			// reset the buffer
			if b == byteTab {
				pieces = dfr.readTabLineAppendPiece(pieces, lineBuffer)
				continue
			}

			// b wasnt a newline, write it to the output buffer.
			lineBuffer.WriteByte(b)
		}
	}

	pieces = dfr.readTabLineAppendPiece(pieces, lineBuffer)
	// we've reached the end of the file
	// there may not have been a newline
	// return what we have
	return pieces, cursor, err
}

// readTabLineAppendPiece is a commonly used code block
// that conditionally adds a new piece to the tab piece collection
func (dfr *DataFileReader) readTabLineAppendPiece(pieces []interface{}, lineBuffer *bytes.Buffer) []interface{} {
	if lineBuffer.Len() > 0 {
		value := lineBuffer.String()
		lineBuffer.Reset()
		if value == null {
			return append(pieces, nil)
		}
		return append(pieces, value)
	}
	return pieces
}
