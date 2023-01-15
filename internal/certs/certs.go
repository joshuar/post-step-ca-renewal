/*
Copyright © 2023 Joshua Rich <joshua.rich@gmail.com>
*/
package certs

type CertConfig struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
	CA   string `json:"ca"`
}
