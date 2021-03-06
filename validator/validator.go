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

package validator

import (
	"context"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	// DefaultFilename is the default filename for the file with the rules of validation
	DefaultFilename = "/data/validation/rules.json"

	// DefaultPluginsDirectory is the default directory where validation plugins are located
	DefaultPluginsDirectory = "/data/validation/"
)

// Config contains the path of the rules JSON file and the directory where the validator scripts are located.
type Config struct {
	RulesPath   string
	PluginsPath string
}

// Validator defines a validator that has an internal state, identified by
// its hash.
type Validator interface {
	// Validate runs validations on a link and returns an error
	// if the link is invalid.
	Validate(context.Context, store.SegmentReader, *cs.Link) error

	// ShouldValidate returns a boolean whether the link should be checked
	ShouldValidate(*cs.Link) bool

	// Hash returns the hash of the validator's state.
	// It can be used to know which set of validations were applied
	// to a block.
	Hash() (*types.Bytes32, error)
}
