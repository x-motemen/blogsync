package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

type WSSEClient struct {
	*http.Client
	UserName string
	Password string
}

func (c *WSSEClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	created := time.Now().Format("2006-01-02T15:04:05Z")

	digest := sha1.New()
	digest.Write(nonce)
	digest.Write([]byte(created))
	digest.Write([]byte(c.Password))

	req.Header.Set(
		"X-WSSE",
		fmt.Sprintf(
			`UsernameToken Username="%s", PasswordDigest="%s", Nonce="%s", Created="%s"`,
			c.UserName,
			base64.StdEncoding.EncodeToString(digest.Sum(nil)),
			base64.StdEncoding.EncodeToString(nonce),
			created,
		),
	)

	return c.Client.Do(req)
}
