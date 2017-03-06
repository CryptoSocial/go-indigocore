// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmstore

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
)

func TestTMStore(t *testing.T) {
	storetestcases.Factory{
		New: func() (store.Adapter, error) {
			config := &Config{
				Endpoint: GetConfig().GetString("rpc_laddr"),
			}
			s := New(config)
			go s.StartWebsocket()
			return s, nil
		},
		Free: func(s store.Adapter) {
			s.(*TMStore).StopWebsocket()
			Reset()
		},
	}.RunTests(t)
}