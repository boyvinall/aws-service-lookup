package ec2tags

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const securityCredentials = "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
const macs = "http://169.254.169.254/latest/meta-data/network/interfaces/macs/"

type RoleCredentials struct {
	Code            string
	LastUpdated     time.Time
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      time.Time
}

func GetList(rawurl string) ([]string, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	list := []string{}
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		list = append(list, s.Text())
	}

	return list, nil
}

func GetDocument(rawurl string) ([]byte, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return b, nil
}

func GetKeysFromRole() (*RoleCredentials, error) {
	credsList, err := GetList(securityCredentials)
	if len(credsList) == 0 {
		return nil, errors.New("No IAM roles found")
	}

	credName := credsList[0]
	b, err := GetDocument(securityCredentials + credName)
	if err != nil {
		return nil, err
	}

	cred := RoleCredentials{}
	err = json.Unmarshal(b, &cred)
	if err != nil {
		return nil, err
	}

	return &cred, nil
}

func GetLocalVPCs() ([]string, error) {
	vpcs := []string{}

	macsList, err := GetList(macs)
	if err != nil {
		return nil, err
	}
	if len(macsList) == 0 {
		return nil, errors.New("unable to get list of MACs")
	}

	for _, m := range macsList {
		vpc, err := GetDocument(macs + m + "/vpc-id")
		if err != nil {
			continue
		}
		vpcs = append(vpcs, string(vpc))
	}

	return vpcs, nil
}
