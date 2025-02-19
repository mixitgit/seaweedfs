package filer

import (
	"bytes"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/wdclient"
	"time"
)

func ReadEntry(masterClient *wdclient.MasterClient, filerClient filer_pb.SeaweedFilerClient, dir, name string, byteBuffer *bytes.Buffer) error {

	request := &filer_pb.LookupDirectoryEntryRequest{
		Directory: dir,
		Name:      name,
	}
	respLookupEntry, err := filer_pb.LookupEntry(filerClient, request)
	if err != nil {
		return err
	}
	if len(respLookupEntry.Entry.Content) > 0 {
		_, err = byteBuffer.Write(respLookupEntry.Entry.Content)
		return err
	}

	return StreamContent(masterClient, byteBuffer, respLookupEntry.Entry.Chunks, 0, int64(FileSize(respLookupEntry.Entry)))

}

func ReadInsideFiler(filerClient filer_pb.SeaweedFilerClient, dir, name string) (content []byte, err error) {
	request := &filer_pb.LookupDirectoryEntryRequest{
		Directory: dir,
		Name:      name,
	}
	respLookupEntry, err := filer_pb.LookupEntry(filerClient, request)
	if err != nil {
		return
	}
	content = respLookupEntry.Entry.Content
	return
}

func SaveInsideFiler(client filer_pb.SeaweedFilerClient, dir, name string, content []byte) error {

	resp, err := filer_pb.LookupEntry(client, &filer_pb.LookupDirectoryEntryRequest{
		Directory: dir,
		Name:      name,
	})

	if err == filer_pb.ErrNotFound {
		err = filer_pb.CreateEntry(client, &filer_pb.CreateEntryRequest{
			Directory: dir,
			Entry: &filer_pb.Entry{
				Name:        name,
				IsDirectory: false,
				Attributes: &filer_pb.FuseAttributes{
					Mtime:       time.Now().Unix(),
					Crtime:      time.Now().Unix(),
					FileMode:    uint32(0644),
					Collection:  "",
					Replication: "",
					FileSize:    uint64(len(content)),
				},
				Content: content,
			},
		})
	} else if err == nil {
		entry := resp.Entry
		entry.Content = content
		entry.Attributes.Mtime = time.Now().Unix()
		entry.Attributes.FileSize = uint64(len(content))
		err = filer_pb.UpdateEntry(client, &filer_pb.UpdateEntryRequest{
			Directory: dir,
			Entry:     entry,
		})
	}

	return err
}
