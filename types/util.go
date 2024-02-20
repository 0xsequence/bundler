package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"

	"github.com/0xsequence/ethkit/go-ethereum/common"

	chunker "github.com/ipfs/boxo/chunker"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func FromHex(str string) ([]byte, error) {
	res := common.FromHex(str)
	if len(res) == 0 && len(str) > 3 {
		return nil, fmt.Errorf("invalid hex string: %s", str)
	}

	return res, nil
}

func HexToBigInt(str string) (*big.Int, error) {
	var ok bool
	var res *big.Int

	// If starts with 0x then it is a hex string
	if len(str) >= 2 && str[:2] == "0x" {
		res, ok = new(big.Int).SetString(str[2:], 16)
	} else {
		res, ok = new(big.Int).SetString(str, 10)
	}

	if !ok {
		return nil, fmt.Errorf("invalid big int string: %s", str)
	}

	return res, nil
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

func ReportToIPFS(ipfsurl string, data []byte) (string, error) {
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
