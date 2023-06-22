package ipfs

import (
	"bytes"
	"io/ioutil"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFSClient struct {
	sh *shell.Shell
}

func NewIPFSClient(apiAddress string) *IPFSClient {
	sh := shell.NewShell(apiAddress)
	return &IPFSClient{sh}
}

func (c *IPFSClient) UploadFile(data []byte) (string, error) {
	cid, err := c.sh.Add(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	return "https://ipfs.io/ipfs/" + cid, nil
}

func (c *IPFSClient) DownloadFile(cid string) ([]byte, error) {
	reader, err := c.sh.Cat(cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *IPFSClient) DeleteFile(cid string) error {
	err := c.sh.Unpin(cid)
	if err != nil {
		return err
	}
	return nil
}
