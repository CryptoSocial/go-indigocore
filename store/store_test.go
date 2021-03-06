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

package store_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	sliceSize = 100
)

var (
	prevLinkHashTestingValue      string
	linkHashTestingValue          string
	badLinkHashTestingValue       string
	emptyPrevLinkHashTestingValue = ""

	segmentSlice cs.SegmentSlice
	stringSlice  []string
)

func init() {
	prevLinkHashTestingValue = testutil.RandomHash().String()
	badLinkHashTestingValue = testutil.RandomHash().String()

	segmentSlice = make(cs.SegmentSlice, sliceSize)
	stringSlice = make([]string, sliceSize)
	for i := 0; i < sliceSize; i++ {
		segmentSlice[i] = cstesting.RandomSegment()
		stringSlice[i] = testutil.RandomString(10)
	}
}

func defaultTestingSegment() *cs.Segment {
	link := &cs.Link{
		Meta: cs.LinkMeta{
			PrevLinkHash: prevLinkHashTestingValue,
			Process:      "TheProcess",
			MapID:        "TheMapId",
			Tags:         []string{"Foo", "Bar"},
			Priority:     42.,
		},
	}
	return link.Segmentify()
}

func emptyPrevLinkHashTestingSegment() *cs.Segment {
	seg := defaultTestingSegment()
	seg.Link.Meta.PrevLinkHash = ""
	return seg
}

func TestSegmentFilter_Match(t *testing.T) {
	type fields struct {
		Pagination   store.Pagination
		MapIDs       []string
		Process      string
		PrevLinkHash *string
		LinkHashes   []string
		Tags         []string
	}
	type args struct {
		segment *cs.Segment
	}
	linkHashesSegment := defaultTestingSegment()
	linkHashesSegmentHash := linkHashesSegment.GetLinkHashString()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Null segment",
			fields: fields{},
			args:   args{nil},
			want:   false,
		},
		{
			name:   "Empty filter",
			fields: fields{},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Good mapId",
			fields: fields{MapIDs: []string{"TheMapId"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad mapId",
			fields: fields{MapIDs: []string{"AMapId"}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Good several mapIds",
			fields: fields{MapIDs: []string{"TheMapId", "SecondMapId"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Good process",
			fields: fields{Process: "TheProcess"},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad process",
			fields: fields{Process: "AProcess"},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Empty prevLinkHash ko",
			fields: fields{PrevLinkHash: &emptyPrevLinkHashTestingValue},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Empty prevLinkHash ok",
			fields: fields{PrevLinkHash: &emptyPrevLinkHashTestingValue},
			args:   args{emptyPrevLinkHashTestingSegment()},
			want:   true,
		},
		{
			name:   "Good prevLinkHash",
			fields: fields{PrevLinkHash: &prevLinkHashTestingValue},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad prevLinkHash",
			fields: fields{PrevLinkHash: &badLinkHashTestingValue},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "LinkHashes ok",
			fields: fields{LinkHashes: []string{testutil.RandomHash().String(), linkHashesSegmentHash}},
			args:   args{linkHashesSegment},
			want:   true,
		},
		{
			name:   "LinkHashes ko",
			fields: fields{LinkHashes: []string{testutil.RandomHash().String()}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "One tag",
			fields: fields{Tags: []string{"Foo"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Two tags",
			fields: fields{Tags: []string{"Foo", "Bar"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Only one good tag",
			fields: fields{Tags: []string{"Foo", "Baz"}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Bad tag",
			fields: fields{Tags: []string{"Hello"}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := store.SegmentFilter{
				Pagination:   tt.fields.Pagination,
				MapIDs:       tt.fields.MapIDs,
				Process:      tt.fields.Process,
				LinkHashes:   tt.fields.LinkHashes,
				PrevLinkHash: tt.fields.PrevLinkHash,
				Tags:         tt.fields.Tags,
			}
			if got := filter.Match(tt.args.segment); got != tt.want {
				t.Errorf("SegmentFilter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapFilter_Match(t *testing.T) {
	type fields struct {
		Pagination store.Pagination
		Process    string
	}
	type args struct {
		segment *cs.Segment
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Null segment",
			fields: fields{},
			args:   args{nil},
			want:   false,
		},
		{
			name:   "Empty filter",
			fields: fields{},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Good process",
			fields: fields{Process: "TheProcess"},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad process",
			fields: fields{Process: "AProcess"},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := store.MapFilter{
				Pagination: tt.fields.Pagination,
				Process:    tt.fields.Process,
			}
			if got := filter.Match(tt.args.segment); got != tt.want {
				t.Errorf("MapFilter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func defaultTestingPagination() store.Pagination {
	return store.Pagination{
		Offset: 0,
		Limit:  10,
	}
}
func TestPagination_PaginateSegments(t *testing.T) {
	type fields struct {
		Offset int
		Limit  int
	}
	type args struct {
		a cs.SegmentSlice
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   cs.SegmentSlice
	}{
		{
			name: "Nothing to paginate",
			fields: fields{
				Offset: 0,
				Limit:  2 * sliceSize,
			},
			args: args{segmentSlice},
			want: segmentSlice,
		},
		{
			name: "Paginate from beginning",
			fields: fields{
				Offset: 0,
				Limit:  sliceSize / 2,
			},
			args: args{segmentSlice},
			want: segmentSlice[:sliceSize/2],
		},
		{
			name: "Paginate from offset",
			fields: fields{
				Offset: 5,
				Limit:  sliceSize / 2,
			},
			args: args{segmentSlice},
			want: segmentSlice[5 : 5+sliceSize/2],
		},
		{
			name: "Paginate zero limit",
			fields: fields{
				Offset: 0,
				Limit:  0,
			},
			args: args{segmentSlice},
			want: cs.SegmentSlice{},
		},
		{
			name: "Paginate outer offset",
			fields: fields{
				Offset: 2 * sliceSize,
				Limit:  sliceSize,
			},
			args: args{segmentSlice},
			want: cs.SegmentSlice{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &store.Pagination{
				Offset: tt.fields.Offset,
				Limit:  tt.fields.Limit,
			}
			if got := p.PaginateSegments(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pagination.PaginateSegments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPagination_PaginateStrings(t *testing.T) {
	type fields struct {
		Offset int
		Limit  int
	}
	type args struct {
		a []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Nothing to paginate",
			fields: fields{
				Offset: 0,
				Limit:  2 * sliceSize,
			},
			args: args{stringSlice},
			want: stringSlice,
		},
		{
			name: "Paginate from beginning",
			fields: fields{
				Offset: 0,
				Limit:  sliceSize / 2,
			},
			args: args{stringSlice},
			want: stringSlice[:sliceSize/2],
		},
		{
			name: "Paginate from offset",
			fields: fields{
				Offset: 5,
				Limit:  sliceSize / 2,
			},
			args: args{stringSlice},
			want: stringSlice[5 : 5+sliceSize/2],
		},
		{
			name: "Paginate zero limit",
			fields: fields{
				Offset: 0,
				Limit:  0,
			},
			args: args{stringSlice},
			want: []string{},
		},
		{
			name: "Paginate outer offset",
			fields: fields{
				Offset: 2 * sliceSize,
				Limit:  sliceSize,
			},
			args: args{stringSlice},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &store.Pagination{
				Offset: tt.fields.Offset,
				Limit:  tt.fields.Limit,
			}
			if got := p.PaginateStrings(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pagination.PaginateStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvents(t *testing.T) {
	t.Run("SavedLinks constructor", func(t *testing.T) {
		e := store.NewSavedLinks()
		assert.EqualValues(t, store.SavedLinks, e.EventType)
		assert.IsType(t, []*cs.Link{}, e.Data, "Event.Data should be a slice of *cs.Link")
	})

	t.Run("Links can be added to SavedLinks event", func(t *testing.T) {
		e := store.NewSavedLinks()
		assert.Empty(t, e.Data, "Links should be initially empty")

		e.AddSavedLinks(cstesting.RandomLink(), cstesting.RandomLink())
		assert.Len(t, e.Data, 2, "Two links should have been added")
	})

	t.Run("SavedLinks event can be initialized with links", func(t *testing.T) {
		e := store.NewSavedLinks(cstesting.RandomLink(), cstesting.RandomLink())
		assert.Len(t, e.Data, 2, "Links should be initially empty")
	})

	t.Run("SavedEvidences constructor", func(t *testing.T) {
		e := store.NewSavedEvidences()
		assert.EqualValues(t, store.SavedEvidences, e.EventType)
		assert.IsType(t, map[string]*cs.Evidence{}, e.Data, "Event.Data should be a map of string/*cs.Evidence")
	})

	t.Run("Evidence can be added to SavedEvidences event", func(t *testing.T) {
		e := store.NewSavedEvidences()
		assert.Empty(t, e.Data, "Evidences should be initially empty")

		linkHash := testutil.RandomHash()
		evidence := cstesting.RandomEvidence()
		e.AddSavedEvidence(linkHash, evidence)

		assert.Len(t, e.Data, 1, "An evidence should have been added")
		evidences := e.Data.(map[string]*cs.Evidence)
		assert.EqualValues(t, evidence, evidences[linkHash.String()], "Invalid evidence")
	})

	t.Run("SavedLinks serialization", func(t *testing.T) {
		link := cstesting.RandomLink()
		e := store.NewSavedLinks(link)

		b, err := json.Marshal(e)
		assert.NoError(t, err)

		var e2 store.Event
		err = json.Unmarshal(b, &e2)
		assert.NoError(t, err)
		assert.EqualValues(t, e.EventType, e2.EventType, "Invalid event type")

		links := e2.Data.([]*cs.Link)
		assert.Len(t, links, 1, "Invalid number of links")
		assert.EqualValues(t, link, links[0], "Invalid link")
	})

	t.Run("SavedEvidences serialization", func(t *testing.T) {
		e := store.NewSavedEvidences()
		evidence := cstesting.RandomEvidence()
		linkHash := testutil.RandomHash()
		e.AddSavedEvidence(linkHash, evidence)

		b, err := json.Marshal(e)
		assert.NoError(t, err)

		var e2 store.Event
		err = json.Unmarshal(b, &e2)
		assert.NoError(t, err)
		assert.EqualValues(t, e.EventType, e2.EventType, "Invalid event type")

		evidences := e2.Data.(map[string]*cs.Evidence)
		deserialized := evidences[linkHash.String()]
		deserialized.Proof = nil
		assert.EqualValues(t, evidence, deserialized, "Invalid evidence")
	})

}
