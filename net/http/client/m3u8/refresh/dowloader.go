package m3u8

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/hopeio/utils/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	tsExt         = ".ts"
	progressWidth = 40
)

type Downloader struct {
	filePath string
	tsDir    string
	finish   int32
	segLen   int
	url      string
}

// NewTask returns a Task instance
func NewTask(filePath, tsFolder string, url string) (*Downloader, error) {
	result, err := FromURL(url)
	if err != nil {
		return nil, err
	}
	// If no output folder specified, use current directory
	if filePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		filePath = pwd + fs.PathSeparator + filePath
	} else {
		if err := os.MkdirAll(filepath.Clean(filepath.Dir(filePath)), os.ModePerm); err != nil {
			return nil, fmt.Errorf("create storage folder failed: %s", err.Error())
		}
	}

	if err := os.MkdirAll(tsFolder, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create ts folder '[%s]' failed: %s", tsFolder, err.Error())
	}
	d := &Downloader{
		filePath: filePath,
		tsDir:    tsFolder,
		url:      url,
	}
	d.segLen = len(result.M3u8.Segments)
	return d, nil
}

func (d *Downloader) SegLen() int {
	return d.segLen
}

// Start runs downloader
func (d *Downloader) Start(concurrency int) error {
	var wg sync.WaitGroup
	wg.Add(d.segLen)
	idxCh := make(chan int, concurrency)
	for _ = range concurrency {
		go func() {
			for idx := range idxCh {
				if err := d.Downloadts(idx); err != nil {
					// Back into the queue, retry request
					fmt.Printf("[failed] %s\n", err.Error())
					<-idxCh
				}
				wg.Done()
			}
		}()
	}
	go func() {
		for i := range d.segLen {
			idxCh <- i
		}
	}()
	wg.Wait()
	close(idxCh)
	return d.Merge()
}

// single thread downloader
func (d *Downloader) Download() error {
	mFile, err := os.Create(d.filePath)
	if err != nil {
		return fmt.Errorf("create main ts file failed：%s", err.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer mFile.Close()

	w := bufio.NewWriter(mFile)
	for segIndex := 0; segIndex < d.segLen; segIndex++ {
		result, err := FromURL(d.url)
		if err != nil {
			return err
		}
		data, err := result.Download(segIndex)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("write to %s: %s", d.filePath, err.Error())
		}
		w.Flush()
	}
	return nil
}

func (d *Downloader) Downloadts(segIndex int) error {
	tsFilename := tsFilename(segIndex)

	fPath := filepath.Join(d.tsDir, tsFilename)

	if fs.NotExist(fPath) {
		result, err := FromURL(d.url)
		if err != nil {
			return err
		}
		data, err := result.Download(segIndex)
		if err != nil {
			return err
		}

		err = fs.Download(fPath, bytes.NewReader(data))
		if err != nil {
			return err
		}
	}
	// Maybe it will be safer in this way...
	atomic.AddInt32(&d.finish, 1)
	//tool.DrawProgressBar("Downloading", float32(d.finish)/float32(d.segLen), progressWidth)
	fmt.Printf("[download %6.2f%%] %s\r", float32(d.finish)/float32(d.segLen)*100, d.url)
	return nil
}

func (d *Downloader) Merge() error {
	// In fact, the number of downloaded segments should be equal to number of m3u8 segments
	missingCount := 0
	for idx := 0; idx < d.segLen; idx++ {
		tsFilename := tsFilename(idx)
		f := filepath.Join(d.tsDir, tsFilename)
		if _, err := os.Stat(f); err != nil {
			missingCount++
		}
	}
	if missingCount > 0 {
		fmt.Printf("[warning] %d files missing\n", missingCount)
	}

	// Create a TS file for merging, all segment files will be written to this file.
	mFile, err := os.Create(d.filePath)
	if err != nil {
		return fmt.Errorf("create main TS file failed：%s", err.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer mFile.Close()

	writer := bufio.NewWriter(mFile)
	mergedCount := 0
	for segIndex := 0; segIndex < d.segLen; segIndex++ {
		tsFilename := tsFilename(segIndex)
		bytes, err := os.ReadFile(filepath.Join(d.tsDir, tsFilename))
		_, err = writer.Write(bytes)
		if err != nil {
			continue
		}
		mergedCount++
		terminal.DrawProgressBar("merge",
			float32(mergedCount)/float32(d.segLen), progressWidth)
	}
	_ = writer.Flush()
	// Remove `ts` folder
	_ = os.RemoveAll(d.tsDir)

	if mergedCount != d.segLen {
		fmt.Printf("[warning] \n%d files merge failed", d.segLen-mergedCount)
	}

	fmt.Printf("\n[output] %s\n", d.filePath)

	return nil
}

func tsFilename(ts int) string {
	return strconv.Itoa(ts) + tsExt
}

func (d *Downloader) FfmpegConcatFile() (string, error) {
	var data bytes.Buffer
	for i := 0; i < d.segLen; i++ {
		data.WriteString(`file '` + d.tsDir + "/" + strconv.Itoa(i) + `.ts'
`)
	}
	ffmpegFilePath := d.tsDir + fs.PathSeparator + "file.txt"

	file, err := os.Create(ffmpegFilePath)
	if err != nil {
		return "", fmt.Errorf("create ffmpeg file failed：%s", err.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()
	_, err = file.Write(data.Bytes())
	if err != nil {
		return "", fmt.Errorf("write to %s: %s", ffmpegFilePath, err.Error())
	}
	return ffmpegFilePath, nil
}

func (d *Downloader) Finished() bool {
	return d.finish == int32(d.segLen)
}

func (d *Downloader) RemoveTmp() error {
	return os.RemoveAll(d.tsDir)
}
