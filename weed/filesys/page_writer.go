package filesys

import (
	"github.com/chrislusf/seaweedfs/weed/filesys/page_writer"
	"github.com/chrislusf/seaweedfs/weed/glog"
)

type PageWriter struct {
	fh            *FileHandle
	collection    string
	replication   string
	chunkSize     int64
	writerPattern *WriterPattern

	randomWriter page_writer.DirtyPages
}

var (
	_ = page_writer.DirtyPages(&PageWriter{})
)

func newPageWriter(fh *FileHandle, chunkSize int64) *PageWriter {
	pw := &PageWriter{
		fh:            fh,
		chunkSize:     chunkSize,
		writerPattern: NewWriterPattern(chunkSize),
		randomWriter:  newMemoryChunkPages(fh, chunkSize),
		// randomWriter: newTempFileDirtyPages(fh.f, chunkSize),
	}
	return pw
}

func (pw *PageWriter) AddPage(offset int64, data []byte) {

	glog.V(4).Infof("%v AddPage [%d, %d) streaming:%v", pw.fh.f.fullpath(), offset, offset+int64(len(data)), pw.writerPattern.IsStreamingMode())

	chunkIndex := offset / pw.chunkSize
	for i := chunkIndex; len(data) > 0; i++ {
		writeSize := min(int64(len(data)), (i+1)*pw.chunkSize-offset)
		pw.addToOneChunk(i, offset, data[:writeSize])
		offset += writeSize
		data = data[writeSize:]
	}
}

func (pw *PageWriter) addToOneChunk(chunkIndex, offset int64, data []byte) {
	pw.randomWriter.AddPage(offset, data)
}

func (pw *PageWriter) FlushData() error {
	pw.writerPattern.Reset()
	return pw.randomWriter.FlushData()
}

func (pw *PageWriter) ReadDirtyDataAt(data []byte, offset int64) (maxStop int64) {
	glog.V(4).Infof("ReadDirtyDataAt %v [%d, %d)", pw.fh.f.fullpath(), offset, offset+int64(len(data)))

	chunkIndex := offset / pw.chunkSize
	for i := chunkIndex; len(data) > 0; i++ {
		readSize := min(int64(len(data)), (i+1)*pw.chunkSize-offset)

		maxStop = pw.randomWriter.ReadDirtyDataAt(data[:readSize], offset)

		offset += readSize
		data = data[readSize:]
	}

	return
}

func (pw *PageWriter) GetStorageOptions() (collection, replication string) {
	return pw.randomWriter.GetStorageOptions()
}

func (pw *PageWriter) LockForRead(startOffset, stopOffset int64) {
	pw.randomWriter.LockForRead(startOffset, stopOffset)
}

func (pw *PageWriter) UnlockForRead(startOffset, stopOffset int64) {
	pw.randomWriter.UnlockForRead(startOffset, stopOffset)
}

func (pw *PageWriter) Destroy() {
	pw.randomWriter.Destroy()
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
