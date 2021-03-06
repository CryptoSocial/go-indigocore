// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgresstore

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/go-indigocore/utils"
)

func TestMain(m *testing.M) {
	const (
		domain = "0.0.0.0"
		port   = "5432"
	)

	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)
	flag.Parse()

	// Postgres container configuration.
	imageName := "postgres:10.1"
	containerName := "indigo_postgresstore_test"
	p, _ := nat.NewPort("tcp", port)
	exposedPorts := map[nat.Port]struct{}{p: {}}
	portBindings := nat.PortMap{
		p: []nat.PortBinding{
			{
				HostIP:   domain,
				HostPort: port,
			},
		},
	}

	// Stop container if it is already running, swallow error.
	utils.KillContainer(containerName)

	// Start postgres container
	if err := utils.RunContainer(containerName, imageName, exposedPorts, portBindings); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	// Retry until container is ready.
	if err := utils.Retry(createDatabase, 10); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	// Run tests.
	testResult := m.Run()

	// Stop postgres container.
	if err := utils.KillContainer(containerName); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	os.Exit(testResult)

}

func createDatabase(attempt int) (bool, error) {
	db, err := sql.Open("postgres", "postgres://postgres@localhost?sslmode=disable")
	if err != nil {
		time.Sleep(1 * time.Second)
		return true, err
	}
	if _, err := db.Exec("CREATE DATABASE sdk_test;"); err != nil {
		time.Sleep(1 * time.Second)
		return true, err
	}
	return false, err
}
