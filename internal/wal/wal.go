package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	WAL_ADD uint16 = 32
	WAL_DEL uint16 = 64
)

const (
	WAL_HEADER_V1 = "MUNCH-01"
)

func i64ToByteArray(i int64) (arr []byte) {
	arr = make([]byte, 8)
	binary.BigEndian.PutUint64(arr[0:8], uint64(i))
	return
}

func ui64ToByteArray(i uint64) (arr []byte) {
	arr = make([]byte, 8)
	binary.BigEndian.PutUint64(arr[0:8], i)
	return
}

func i32ToByteArray(i int32) (arr []byte) {
	arr = make([]byte, 4)
	binary.BigEndian.PutUint32(arr[0:4], uint32(i))
	return
}

func i16ToByteArray(i int16) (arr []byte) {
	arr = make([]byte, 2)
	binary.BigEndian.PutUint16(arr[0:2], uint16(i))
	return
}

func ui16ToByteArray(i uint16) (arr []byte) {
	arr = make([]byte, 2)
	binary.BigEndian.PutUint16(arr[0:2], i)
	return
}

func FindFilesOnOrAfter(walDir, walPrefix string, timestamp uint64) ([]string, error) {
	s, err := os.Stat(walDir)
	if err != nil {
		return []string{}, err
	}
	if s.IsDir() == false {
		return []string{}, errors.New(walDir + " is not a directory!")
	}
	entries, err := os.ReadDir(walDir)
	if err != nil {
		return []string{}, err
	}

	var files []string = make([]string, 0)
	for _, v := range entries {
		if !v.IsDir() && strings.HasPrefix(v.Name(), walPrefix) {
			files = append(files, v.Name())
		}
		sort.Strings(files)
	}
	if len(files) == 0 {
		return []string{}, nil
	}
	var tsBefore int64 = 0
	var fileTs []int64 = make([]int64, 0)
	for _, fname := range files {
		ts := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(fname), ".wal"), walPrefix)
		tsint, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return []string{}, err
		}
		if tsint > tsBefore && tsint <= int64(timestamp) {
			tsBefore = tsint
			fileTs = append(fileTs, tsint)
		}
	}
	var fileList []string = make([]string, 0)
	for _, f := range fileTs {
		fileList = append(files, filepath.Join(walDir, (walPrefix+strconv.FormatInt(f, 10)+".wal")))
	}
	return fileList, nil
}

// WalEntry represents a single entry in the WAL file.
type WalEntry struct {
	Timestamp  uint64
	KeyLen     uint16
	PatternLen uint16
	Key        []byte
	Pattern    []byte
	Action     uint16
}

func (we *WalEntry) GetPatternAsJson() (string, error) {
	return "", nil
}

// WalFile is responsible for managing file state and read/writes.
type WalFile struct {
	FileName string
	file     *os.File
	size     int64
	entry    int
	currByte int64
	mu       sync.Mutex
	isClosed bool
}

// CreateWalFile creates a new WAL file and generates a header.
func CreateWalFile(dirName, filePrefix string) (string, error) {
	ts := time.Now().UnixNano()
	fileName := filepath.Join(dirName, (filePrefix + strconv.FormatInt(ts, 10) + ".wal"))
	_, err := os.Stat(fileName)
	if err != nil {
		f, err := os.Create(fileName)
		if err != nil {
			return fileName, err
		}
		_, err = f.Write([]byte(WAL_HEADER_V1))
		if err != nil {
			return fileName, err
		}
		for x := 0; x < 248; x++ {
			_, err = f.Write([]byte{0x00})
			if err != nil {
				return fileName, err
			}
		}
		_ = f.Sync()
		_, _ = f.Seek(0, 0)
		return fileName, nil
	}
	return fileName, os.ErrExist
}

// OpenWalFile opens an existing WAL file created with CreateWalFile()
func OpenWalFile(fileName string) (*WalFile, error) {
	s, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return nil, errors.New("File is directory!")
	}
	f, err := os.OpenFile(fileName, os.O_RDWR, s.Mode())
	if err != nil {
		return nil, err
	}

	// check to see if it has a header
	hdrBuf := make([]byte, 256)
	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	rCt, err := f.Read(hdrBuf)
	if err != nil {
		return nil, err
	}
	if rCt < len(hdrBuf) {
		return nil, errors.New(fmt.Sprintf("Sought to read %d bytes, read %d", len(hdrBuf), rCt))
	}

	if string(hdrBuf[0:8]) != WAL_HEADER_V1 {
		return nil, errors.New(fmt.Sprintf("Incorrect WAL header (got \"%s\"" + string(hdrBuf[0:8])))
	}
	for x := 0; x < 248; x++ {
		if hdrBuf[8+x] != 0x0 {
			return nil, errors.New(fmt.Sprintf("Header byte %d not null!", (x + 8)))
		}
	}
	wf := &WalFile{
		FileName: fileName,
		file:     f,
		mu:       sync.Mutex{},
		size:     s.Size(),
	}
	_, err = wf.file.Seek(256, 0)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

// Close closes the WAL file.
func (wf *WalFile) Close() {
	wf.currByte = 0
	wf.entry = 0
	_ = wf.file.Sync()
	_ = wf.file.Close()
}

// Rewind rewinds a WAL file back to the end of the file header.
func (wf *WalFile) Rewind() {
	wf.currByte = 0
	wf.entry = 0
	_, _ = wf.file.Seek(256, 0)
	_ = wf.file.Sync()
}

// Delete removes a WAL file entirely.
func (wf *WalFile) Delete() error {
	err := os.Remove(wf.FileName)
	if err != nil {
		return err
	}
	wf.file = nil
	wf.currByte = 0
	wf.entry = 0
	return nil
}

func (wf *WalFile) OnOrAfter(timestamp uint64) (*WalEntry, error) {
	wf.Rewind()
	for wf.HasNext() {
		we, err := wf.Next()
		if err != nil {
			return nil, err
		}
		if we.Timestamp >= timestamp {
			wf.Rewind()
			return &we, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("No records equal or older than %d found", timestamp))
}

// HasNext indicates whether there are additional WAL entries in
// the current file.
func (wf *WalFile) HasNext() bool {
	return wf.size > (wf.currByte + 256)
}

// Next retrieves the next WAL entry from the current file.
func (wf *WalFile) Next() (WalEntry, error) {
	wf.mu.Lock()
	defer wf.mu.Unlock()
	var bytesRead = 0
	var timestamp uint64
	var keyLen, patLen uint16
	var headerBuf []byte = make([]byte, 13)
	ct, err := wf.file.Read(headerBuf)
	if err != nil {
		return WalEntry{}, err
	}

	bytesRead = ct

	if headerBuf[0] != 0x00 {
		return WalEntry{}, errors.New(fmt.Sprintf("Invalid header byte (Got '%s')", string(headerBuf[0])))
	}

	timestamp = binary.BigEndian.Uint64(headerBuf[1:9])
	keyLen = binary.BigEndian.Uint16(headerBuf[9:11])
	patLen = binary.BigEndian.Uint16(headerBuf[11:])

	var keyBuf []byte = make([]byte, keyLen)
	ct, err = wf.file.Read(keyBuf)
	if err != nil {
		return WalEntry{}, err
	}
	bytesRead = bytesRead + ct

	var patBuf []byte = make([]byte, patLen)
	ct, err = wf.file.Read(patBuf)
	if err != nil {
		return WalEntry{}, err
	}
	bytesRead = bytesRead + ct

	var actVal uint16
	actBuf := make([]byte, 2)
	ct, err = wf.file.Read(actBuf)
	if err != nil {
		return WalEntry{}, err
	}
	actVal = binary.BigEndian.Uint16(actBuf)
	bytesRead = bytesRead + ct
	wf.currByte = wf.currByte + int64(bytesRead)
	wf.entry++
	w := WalEntry{
		Timestamp:  timestamp,
		KeyLen:     keyLen,
		PatternLen: patLen,
		Key:        keyBuf,
		Pattern:    patBuf,
		Action:     actVal,
	}
	return w, nil
}

func writeWalEntryHeader(f *os.File) error {
	return nil
}

// Write writes a WAL entry to the current file.
func (wf *WalFile) Write(timestamp int64, key, pattern []byte, act uint16) error {
	if wf.file == nil {
		return os.ErrNotExist
	}
	wf.mu.Lock()
	defer wf.mu.Unlock()

	_, err := wf.file.Seek(0, 2)
	if err != nil {
		return err
	}
	_, err = wf.file.Write([]byte{0x0})
	if err != nil {
		return err
	}
	wf.size++
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}
	// Write timestamp
	_, err = wf.file.Write(i64ToByteArray(timestamp))
	if err != nil {
		return err
	}
	wf.size = wf.size + int64(8)
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}

	// Write key size
	var keySz uint16 = uint16(len(key))
	_, err = wf.file.Write(ui16ToByteArray(keySz))
	if err != nil {
		return err
	}
	wf.size = wf.size + int64(2)
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}

	// Write pattern size
	var patSz uint16 = uint16(len(pattern))
	_, err = wf.file.Write(ui16ToByteArray(patSz))
	if err != nil {
		return err
	}
	wf.size = wf.size + int64(2)
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}

	// Write key
	_, err = wf.file.Write(key)
	if err != nil {
		return err
	}
	wf.size = wf.size + int64(keySz)
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}

	// Write pattern
	_, err = wf.file.Write(pattern)
	if err != nil {
		return err
	}
	wf.size = wf.size + int64(patSz)
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}

	// Write action flag
	_, err = wf.file.Write(ui16ToByteArray(act))
	if err != nil {
		return err
	}
	wf.size = wf.size + int64(2)
	_, err = wf.file.Seek(0, 2)
	if err != nil {
		return err
	}

	_ = wf.file.Sync()
	return nil
}
