package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	chunker "github.com/ipfs/boxo/chunker"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

type Client string

func NewClient(url string) *Client {
	if url == "" {
		return nil
	}

	c := Client(url)
	return &c
}

func (ipfs *Client) ReportToIPFS(data []byte) (string, error) {
	if ipfs == nil || *ipfs == "" {
		return "", fmt.Errorf("ipfs url not set")
	}

	ipfsurl := string(*ipfs)
	if ipfsurl[len(ipfsurl)-1] != '/' {
		ipfsurl += "/"
	}

	url := ipfsurl + "api/v0/add?cid-version=1"

	// Prepare the file to upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "data")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	// Create the request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res struct {
		Hash string `json:"Hash"`
	}

	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return "", err
	}

	return res.Hash, nil
}

func Cid(data []byte) (string, error) {
	// Create an IPLD UnixFS chunker with size 1 MiB
	chunks := chunker.NewSizeSplitter(bytes.NewReader(data), 1024*1024)

	// Concatenate the chunks to build the DAG
	var buf bytes.Buffer
	for {
		chunk, err := chunks.NextBytes()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		buf.Write(chunk)
	}

	// Calculate the CID for the DAG
	hash, err := mh.Sum(data, mh.SHA2_256, -1)
	if err != nil {
		return "", err
	}

	// Create a CID version 1 (with multibase encoding base58btc)
	c := cid.NewCidV1(cid.Raw, hash)

	// Print the CID as a string
	return c.String(), nil
}

func IsCid(str string) bool {
	_, err := cid.Decode(str)
	return err == nil
}
