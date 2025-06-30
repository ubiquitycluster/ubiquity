package main

/*
# Copyright The Ubiquity Authors.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.
*/
// TODO WIP clean up

import (
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v2"
)

type RandomPassword struct {
	Path string
	Data []struct {
		Key     string
		Length  int
		Special bool
	}
}

func main() {
	data, err := os.ReadFile("./config.yaml")

	if err != nil {
		log.Fatalf("unable to read config file: %v", err)
	}

	randomPasswords := []RandomPassword{}

	err = yaml.Unmarshal([]byte(data), &randomPasswords)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	config := vault.DefaultConfig()

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	for _, randomPassword := range randomPasswords {
		path := fmt.Sprintf("/secret/data/%s", randomPassword.Path)

		secret, _ := client.Logical().Read(path)

		if secret == nil {
			secretData := map[string]interface{}{
				"data": map[string]interface{}{},
			}

			for _, randomKey := range randomPassword.Data {
				res, err := password.Generate(32, 3, 3, false, true)
				if err != nil {
					log.Fatal(err)
				}

				secretData["data"].(map[string]interface{})[randomKey.Key] = res
			}

			_, err = client.Logical().Write(path, secretData)

			if err != nil {
				log.Fatalf("Unable to write secret: %v", err)
			} else {
				log.Println("Secret written successfully.")
			}
		} else {
			log.Println("Key abc in secret already existed.")
		}
	}
}
