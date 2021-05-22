package login

import (
	"encoding/json"
	"strings"

	"github.com/chanbakjsd/gotrix"
	"github.com/chanbakjsd/gotrix/matrix"
	"github.com/zalando/go-keyring"
)

type session struct { // TODO: scheme
	Homeserver  string
	UserID      matrix.UserID
	DeviceID    matrix.DeviceID
	AccessToken string
}

func saveSession(c *gotrix.Client) error {
	var b strings.Builder
	json.NewEncoder(&b).Encode(session{
		Homeserver:  c.HomeServer,
		UserID:      c.UserID,
		DeviceID:    c.DeviceID,
		AccessToken: c.AccessToken,
	})

	return keyring.Set("matrixemopicker", "_main", b.String())
}

func hasSavedSession() bool {
	_, err := keyring.Get("matrixemopicker", "_main")
	return err == nil
}

func restoreSession() (*gotrix.Client, error) {
	data, err := keyring.Get("matrixemopicker", "_main")
	if err != nil {
		return nil, err
	}

	var ses session
	if err := json.NewDecoder(strings.NewReader(data)).Decode(&ses); err != nil {
		return nil, err
	}

	client, err := gotrix.New(ses.Homeserver)
	if err != nil {
		return nil, err
	}

	client.UserID = ses.UserID
	client.DeviceID = ses.DeviceID
	client.AccessToken = ses.AccessToken

	return client, nil
}
