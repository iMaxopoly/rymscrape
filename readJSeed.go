package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"io/ioutil"
)

// ENCKEY is the XOR encrypted-key that we use to encrypt and decrypt the JSeed files.
// Real Value: "&P~qZ|;'f|\\mznLom\"d^|SgE&DDu,l;E"
const ENCKEY = `&P~qZ|;'f|\mznLom"d^|SgE&DDu,l;E`

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
	VideoListLinks []struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"videoListLinks"`
	VideoListAcquire struct {
		LookFor []string `json:"lookFor"`
	} `json:"videoListAcquire"`
}

// DecryptAESGCM decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func DecryptAESGCM(ciphertext []byte, key []byte) ([]byte, error) {
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

// EncryptDecryptXOR runs a XOR encryption on the input string, encrypting it if it hasn't already been,
// and decrypting it if it has, using the key provided.
func EncryptDecryptXOR(input, key string) (output string) {
	for i := 0; i < len(input); i++ {
		output += string(input[i] ^ key[i%len(key)])
	}

	return output
}

// ReadJSeedFile reads the seed file, decrypts it and unmarshals it to an instance of JSeed Struct.
func ReadJSeedFile() JSeed {
	raw, err := ioutil.ReadFile(*confSeedFile)
	handleErrorAndPanic(err)

	raw, err = DecryptAESGCM(raw, []byte(ENCKEY))
	handleErrorAndPanic(err)

	var j JSeed

	err = json.Unmarshal(raw, &j)
	handleErrorAndPanic(err)

	return j
}
