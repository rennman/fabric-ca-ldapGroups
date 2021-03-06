/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/cloudflare/cfssl/config"
)

const (
	rootPort         = 7075
	rootDir          = "rootDir"
	intermediatePort = 7076
	intermediateDir  = "intDir"
	testdataDir      = "../testdata"
)

func getRootServerURL() string {
	return fmt.Sprintf("http://admin:adminpw@localhost:%d", rootPort)
}

// TestGetRootServer creates a server with root configuration
func TestGetRootServer(t *testing.T) *Server {
	return TestGetServer(rootPort, rootDir, "", -1, t)
}

// TestGetIntermediateServer creates a server with intermediate server configuration
func TestGetIntermediateServer(idx int, t *testing.T) *Server {
	return TestGetServer(
		intermediatePort,
		path.Join(intermediateDir, strconv.Itoa(idx)),
		getRootServerURL(),
		-1,
		t)
}

// TestGetServer creates and returns a pointer to a server struct
func TestGetServer(port int, home, parentURL string, maxEnroll int, t *testing.T) *Server {
	return TestGetServer2(home != testdataDir, port, home, parentURL, maxEnroll, t)
}

// TestGetServer2 creates and returns a pointer to a server struct, with an option of
// whether or not to remove the home directory first
func TestGetServer2(deleteHome bool, port int, home, parentURL string, maxEnroll int, t *testing.T) *Server {
	if deleteHome {
		os.RemoveAll(home)
	}
	affiliations := map[string]interface{}{
		"hyperledger": map[string]interface{}{
			"fabric":    []string{"ledger", "orderer", "security"},
			"fabric-ca": nil,
			"sdk":       nil,
		},
		"org2": nil,
	}
	profiles := map[string]*config.SigningProfile{
		"tls": &config.SigningProfile{
			Usage:        []string{"signing", "key encipherment", "server auth", "client auth", "key agreement"},
			ExpiryString: "8760h",
		},
		"ca": &config.SigningProfile{
			Usage:        []string{"cert sign", "crl sign"},
			ExpiryString: "8760h",
			CAConstraint: config.CAConstraint{
				IsCA:       true,
				MaxPathLen: 0,
			},
		},
	}
	defaultProfile := &config.SigningProfile{
		Usage:        []string{"cert sign"},
		ExpiryString: "8760h",
	}
	srv := &Server{
		Config: &ServerConfig{
			Port:  port,
			Debug: true,
		},
		CA: CA{
			Config: &CAConfig{
				Intermediate: IntermediateCA{
					ParentServer: ParentServer{
						URL: parentURL,
					},
				},
				Affiliations: affiliations,
				Registry: CAConfigRegistry{
					MaxEnrollments: maxEnroll,
				},
				Signing: &config.Signing{
					Profiles: profiles,
					Default:  defaultProfile,
				},
			},
		},
		HomeDir: home,
	}
	// The bootstrap user's affiliation is the empty string, which
	// means the user is at the affiliation root
	err := srv.RegisterBootstrapUser("admin", "adminpw", "")
	if err != nil {
		t.Errorf("Failed to register bootstrap user: %s", err)
		return nil
	}
	return srv
}

// CopyFile copies a file
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}
	return nil
}
