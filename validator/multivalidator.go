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
	"crypto/sha256"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

type multiValidator struct {
	validators []Validator
}

// NewMultiValidator creates a validator that will simply be a collection
// of single-purpose validators.
// The slice of validators should be loaded from a JSON file via validator.LoadConfig().
func NewMultiValidator(validators []Validator) Validator {
	return &multiValidator{validators}
}

func (v multiValidator) ShouldValidate(link *cs.Link) bool {
	return true
}

func (v multiValidator) Hash() (*types.Bytes32, error) {
	b := make([]byte, 0)
	for _, validator := range v.validators {
		validatorHash, err := validator.Hash()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		b = append(b, validatorHash[:]...)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

func (v multiValidator) matchValidators(l *cs.Link) (linkValidators []Validator) {
	for _, child := range v.validators {
		if child.ShouldValidate(l) {
			linkValidators = append(linkValidators, child)
		}
	}
	return
}

// Validate runs the validation on every child validator matching the provided link.
// It is the multiValidator's responsability to call child.ShouldValidate() before running the validation.
func (v multiValidator) Validate(ctx context.Context, r store.SegmentReader, l *cs.Link) error {
	linkValidators := v.matchValidators(l)
	if len(linkValidators) == 0 {
		return errors.Errorf("Validation failed: link with process: [%s] and type: [%s] does not match any validator", l.Meta.Process, l.Meta.Type)
	}

	for _, child := range linkValidators {
		err := child.Validate(ctx, r, l)
		if err != nil {
			return err
		}
	}
	return nil
}
