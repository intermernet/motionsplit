package motionsplit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	splitDelimiter = "MotionPhoto_Data"
	chunkSize      = 4096
)

// Split splits a file into a JPEG and an MP4
func Split(file string) error {
	basePrefix := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	fSize := stat.Size()
	b := make([]byte, 2)
	_, err = f.Read(b)
	if err != nil {
		return err
	}
	jpg := []byte{0xFF, 0xD8}
	if !bytes.Equal(b, jpg) {
		return fmt.Errorf("%s: Not a JPEG file", file)
	}
	_, err = f.Seek(io.SeekStart, 0)
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	offset, err := find(f, []byte(splitDelimiter))
	if offset < 0 {
		return fmt.Errorf("%s: Not a motion photo", file)
	}
	_, err = f.Seek(io.SeekStart, 0)
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	jpegFile, err := os.OpenFile(basePrefix+"_photo.jpg", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	defer jpegFile.Close()
	mp4File, err := os.OpenFile(basePrefix+"_video.mp4", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	defer mp4File.Close()
	f.Seek(io.SeekStart, 0)
	written, err := io.CopyN(jpegFile, f, offset)
	if err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	if written != offset {
		return fmt.Errorf("%s: offset and written not equal", file)
	}
	buf := make([]byte, int(fSize)-(int(offset)+len(splitDelimiter))-1)
	n, err := f.ReadAt(buf, offset+1+int64(len(splitDelimiter)))
	if err != nil && err != io.EOF {
		return fmt.Errorf("%s: %v", file, err)
	}
	if _, err := mp4File.Write(buf[:n]); err != nil {
		return fmt.Errorf("%s: %v", file, err)
	}
	return nil
}

func find(r io.Reader, search []byte) (int64, error) {
	var offset int64
	tailLen := len(search) - 1
	chunk := make([]byte, chunkSize+tailLen)
	n, err := r.Read(chunk[tailLen:])

	idx := bytes.Index(chunk[tailLen:n+tailLen], search)
	for {
		if idx >= 0 {
			return offset + int64(idx) - int64(len(splitDelimiter)), nil
		}
		if err == io.EOF {
			return -1, nil
		} else if err != nil {
			return -1, err
		}
		copy(chunk, chunk[chunkSize:])
		offset += chunkSize
		n, err = r.Read(chunk[tailLen:])
		idx = bytes.Index(chunk[:n+tailLen], search)
	}
}
