package models

import (
	"encoding/json"
)

type Profile struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

func NewProfileFromResponse(response []byte) (profile *Profile, err error) {
	profile = &Profile{}
	if err := json.Unmarshal(response, &profile); err != nil {
		return nil, err
	}

	return profile, err
}
