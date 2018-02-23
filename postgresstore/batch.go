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

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// Batch is the type that implements github.com/stratumn/go-indigocore/store.Batch.
type Batch struct {
	*reader
	*writer

	Links      []*cs.Link
	eventChans []chan *store.Event
	done       bool
	tx         *sql.Tx
}

// NewBatch creates a new instance of a Postgres Batch.
func NewBatch(tx *sql.Tx, eventChans []chan *store.Event) (*Batch, error) {
	stmts, err := newBatchStmts(tx)
	if err != nil {
		return nil, err
	}

	return &Batch{
		reader:     &reader{stmts: readStmts(stmts.readStmts)},
		writer:     &writer{stmts: writeStmts(stmts.writeStmts)},
		tx:         tx,
		eventChans: eventChans,
	}, nil
}

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (b *Batch) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	b.Links = append(b.Links, link)
	return b.writer.CreateLink(link)
}

// Write implements github.com/stratumn/go-indigocore/store.Batch.Write.
func (b *Batch) Write() error {
	b.done = true
	if err := b.tx.Commit(); err != nil {
		return err
	}

	for _, c := range b.eventChans {
		c <- store.NewSavedLinks(b.Links...)
	}

	return nil
}
