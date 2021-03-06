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
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseConfig(t *testing.T) {
	process := "p1"
	linkType := "sell"

	type testCaseCfg struct {
		id            string
		process       string
		linkType      string
		schema        []byte
		valid         bool
		expectedError error
	}

	testCases := []testCaseCfg{{
		id:            "missing-process",
		process:       "",
		linkType:      linkType,
		valid:         false,
		expectedError: ErrMissingProcess,
	}, {
		id:            "missing-link-type",
		process:       process,
		linkType:      "",
		valid:         false,
		expectedError: ErrMissingLinkType,
	}, {
		id:       "valid-config",
		process:  process,
		linkType: linkType,
		valid:    true,
	},
	}

	for _, tt := range testCases {
		t.Run(tt.id, func(t *testing.T) {
			cfg, err := newValidatorBaseConfig(
				tt.process,
				tt.linkType,
			)

			if tt.valid {
				assert.NotNil(t, cfg)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, cfg)
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
			}
		})
	}
}

func TestBaseConfig_ShouldValidate(t *testing.T) {
	process := "p1"
	linkType := "sell"
	cfg, err := newValidatorBaseConfig(
		process,
		linkType,
	)
	require.NoError(t, err)

	type testCase struct {
		name           string
		link           *cs.Link
		shouldValidate bool
	}

	testCases := []testCase{
		{
			name:           "valid-link",
			shouldValidate: true,
			link:           cstesting.NewLinkBuilder().WithProcess(process).WithType(linkType).Sign().Build(),
		},
		{
			name:           "no-process",
			shouldValidate: false,
			link:           cstesting.NewLinkBuilder().WithProcess("").WithType(linkType).Sign().Build(),
		},
		{
			name:           "process-no-match",
			shouldValidate: false,
			link:           cstesting.NewLinkBuilder().WithProcess("test").WithType(linkType).Sign().Build(),
		},
		{
			name:           "no-type",
			shouldValidate: false,
			link:           cstesting.NewLinkBuilder().WithProcess(process).WithType("").Sign().Build(),
		},
		{
			name:           "type-no-match",
			shouldValidate: false,
			link:           cstesting.NewLinkBuilder().WithProcess(process).WithType("test").Sign().Build(),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res := cfg.ShouldValidate(tt.link)
			assert.Equal(t, tt.shouldValidate, res)
		})
	}
}
