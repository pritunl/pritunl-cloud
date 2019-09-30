package link

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	transport = &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
	}
	ClientInsec = &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	ClientSec = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 10 * time.Second,
	}
	Hash = ""
)

func decResp(secret, iv, sig, encData string) (cipData []byte, err error) {
	hashFunc := hmac.New(sha512.New, []byte(secret))
	hashFunc.Write([]byte(encData))
	rawSignature := hashFunc.Sum(nil)
	testSig := base64.StdEncoding.EncodeToString(rawSignature)
	if sig != testSig {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Cipher data signature invalid"),
		}
		return
	}

	cipIv, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Failed to decode cipher IV"),
		}
		return
	}

	encKeyHash := sha256.New()
	encKeyHash.Write([]byte(secret))
	cipKey := encKeyHash.Sum(nil)

	cipData, err = base64.StdEncoding.DecodeString(encData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Failed to decode response data"),
		}
		return
	}

	if len(cipIv) != aes.BlockSize {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Invalid cipher key"),
		}
		return
	}

	if len(cipData)%aes.BlockSize != 0 {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Invalid cipher data"),
		}
		return
	}

	block, err := aes.NewCipher(cipKey)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Failed to load cipher"),
		}
		return
	}

	mode := cipher.NewCBCDecrypter(block, cipIv)
	mode.CryptBlocks(cipData, cipData)

	cipData = bytes.TrimRight(cipData, "\x00")

	return
}
