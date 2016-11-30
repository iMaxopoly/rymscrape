package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"io/ioutil"
)

const ENCKEY = "&P~qZ|;'f|\\mznLom\"d^|SgE&DDu,l;E"

// JSeed is a struct that copies the json AES-512 encrypted file contents that describes the operations required to get links from.
type JSeed struct {
	SiteLink        string   `json:"siteLink"`
	SiteProtocol    string   `json:"siteProtocol"`
	SiteSignature   string   `json:"siteSignature"`
	FullListLinks   []string `json:"fullListLinks"`
	FullListAcquire struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"fullListAcquire"`
	EpisodeListAcquire struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"episodeListAcquire"`
	VideoListAcquire struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"videoListAcquire"`
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

// ReadJSeedFile reads the seed file, decrypts it and unmarshals it to an instance of JSeed Struct.
func ReadJSeedFile() JSeed {
	raw, err := ioutil.ReadFile(*confSeedFile)
	handleErrorAndPanic(err)

	raw, err = Decrypt(raw, ENCKEY)
	handleErrorAndPanic(err)

	var j JSeed

	err = json.Unmarshal(raw, &j)
	handleErrorAndPanic(err)

	return j
}
