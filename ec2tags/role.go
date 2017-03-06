package ec2tags

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const securityCredentials = "http://169.254.169.254/latest/meta-data/iam/security-credentials/"

type RoleCredentials struct {
	Code            string
	LastUpdated     time.Time
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      time.Time
}

func GetKeysFromRole() (*RoleCredentials, error) {
	resp, err := http.Get(securityCredentials)
	if err != nil {
		fmt.Printf("get failed: %s", err.Error())
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		fmt.Printf("readfull: %s", err.Error())
		return nil, err
	}

	credsList := []string{}
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		credsList = append(credsList, s.Text())
	}

	if len(credsList) == 0 {
		return nil, errors.New("No IAM roles found")
	}

	credName := credsList[0]
	resp, err = http.Get(securityCredentials + credName)
	if err != nil {
		fmt.Printf("get failed: %s", err.Error())
		return nil, err
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		fmt.Printf("readfull: %s", err.Error())
		return nil, err
	}

	cred := RoleCredentials{}
	err = json.Unmarshal(b, &cred)
	if err != nil {
		return nil, err
	}

	return &cred, nil
}
