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

package tmpoptestcases

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stratumn/go-indigocore/tmpop/tmpoptestcases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckTx tests what happens when the ABCI method CheckTx() is called
func (f Factory) TestCheckTx(t *testing.T) {
	h, _ := f.newTMPop(t, nil)
	defer f.free()

	t.Run("Check valid link returns ok", func(t *testing.T) {
		_, tx := makeCreateRandomLinkTx(t)
		res := h.CheckTx(tx)
		assert.True(t, res.IsOK(), "Expected CheckTx to return an OK result, got %v", res)
	})

	t.Run("Check link with invalid reference returns not-ok", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta.Refs = []cs.SegmentReference{cs.SegmentReference{
			Process:  "proc",
			LinkHash: "invalidLinkHash",
		}}
		tx := makeCreateLinkTx(t, link)

		res := h.CheckTx(tx)

		assert.EqualValues(t, tmpop.CodeTypeValidation, res.Code)
	})

	t.Run("Check link with uncommitted but checked reference returns ok", func(t *testing.T) {
		link, tx := makeCreateRandomLinkTx(t)
		linkHash, _ := link.Hash()
		res := h.CheckTx(tx)

		linkWithRef := cstesting.NewLinkBuilder().WithProcess(link.Meta.Process).Build()
		linkWithRef.Meta.Refs = []cs.SegmentReference{cs.SegmentReference{
			Process:  link.Meta.Process,
			LinkHash: linkHash.String(),
		}}
		tx = makeCreateLinkTx(t, linkWithRef)

		res = h.CheckTx(tx)

		assert.True(t, res.IsOK(), "Expected CheckTx to return an OK result, got %v", res)
	})
}

// TestDeliverTx tests what happens when the ABCI method DeliverTx() is called
func (f Factory) TestDeliverTx(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	h.BeginBlock(req)

	t.Run("Deliver valid link returns ok", func(t *testing.T) {
		_, tx := makeCreateRandomLinkTx(t)
		res := h.DeliverTx(tx)

		assert.True(t, res.IsOK(), "Expected DeliverTx to return an OK result, got %v", res)
	})

	t.Run("Deliver link referencing a checked but not delivered link returns an error", func(t *testing.T) {
		link, tx := makeCreateRandomLinkTx(t)
		linkHash, _ := link.Hash()
		h.CheckTx(tx)

		linkWithRef := cstesting.NewLinkBuilder().WithProcess(link.Meta.Process).Build()
		linkWithRef.Meta.Refs = []cs.SegmentReference{cs.SegmentReference{
			Process:  link.Meta.Process,
			LinkHash: linkHash.String(),
		}}
		tx = makeCreateLinkTx(t, linkWithRef)
		res := h.DeliverTx(tx)

		assert.EqualValues(t, tmpop.CodeTypeValidation, res.Code)
	})
}

// TestCommitTx tests what happens when the ABCI method CommitTx() is called
func (f Factory) TestCommitTx(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	ctrl := gomock.NewController(t)
	tmClientMock := tmpoptestcasesmocks.NewMockTendermintClient(ctrl)
	tmClientMock.EXPECT().Block(gomock.Any(), gomock.Any()).Return(&tmpop.Block{}, nil).AnyTimes()
	h.ConnectTendermint(tmClientMock)

	previousAppHash := req.Header.AppHash
	h.BeginBlock(req)

	link1, tx := makeCreateRandomLinkTx(t)
	h.DeliverTx(tx)

	link2, tx := makeCreateRandomLinkTx(t)
	h.DeliverTx(tx)

	res := h.Commit()
	if len(res.GetData()) == 0 {
		t.Fatalf("Commit failed")
	}

	t.Run("Commit correctly saves links and updates app hash", func(t *testing.T) {
		verifyLinkStored(t, h, link1)
		verifyLinkStored(t, h, link2)

		if bytes.Compare(previousAppHash, res.Data) == 0 {
			t.Errorf("Committed app hash is the same as the previous app hash")
		}
	})

	t.Run("Committed link events are saved and can be queried", func(t *testing.T) {
		var events []*store.Event
		err := makeQuery(h, tmpop.PendingEvents, nil, &events)
		assert.NoError(t, err)
		require.Len(t, events, 1, "Invalid number of events")

		savedEvent := events[0]
		assert.EqualValues(t, store.SavedLinks, savedEvent.EventType)

		savedLinks := savedEvent.Data.([]*cs.Link)
		require.Len(t, savedLinks, 2, "Invalid number of links")
		assert.EqualValues(t, link1, savedLinks[0])
		assert.EqualValues(t, link2, savedLinks[1])
	})
}
