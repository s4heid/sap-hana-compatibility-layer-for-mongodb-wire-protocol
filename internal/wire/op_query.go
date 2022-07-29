// SPDX-FileCopyrightText: 2021 FerretDB Inc.
//
// SPDX-License-Identifier: Apache-2.0

// Copyright 2021 FerretDB Inc.
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

package wire

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"

	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/bson"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/types"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/lazyerrors"
)

// OpQuery is used to query the database for documents in a collection.
type OpQuery struct {
	Flags                OpQueryFlags
	FullCollectionName   string
	NumberToSkip         int32
	NumberToReturn       int32
	Query                types.Document
	ReturnFieldsSelector *types.Document
}

func (query *OpQuery) msgbody() {}

func (query *OpQuery) readFrom(bufr *bufio.Reader) error {
	if err := binary.Read(bufr, binary.LittleEndian, &query.Flags); err != nil {
		return lazyerrors.Errorf("wire.OpQuery.ReadFrom (binary.Read): %w", err)
	}

	var col bson.CString
	if err := col.ReadFrom(bufr); err != nil {
		return err
	}
	query.FullCollectionName = string(col)

	if err := binary.Read(bufr, binary.LittleEndian, &query.NumberToSkip); err != nil {
		return err
	}
	if err := binary.Read(bufr, binary.LittleEndian, &query.NumberToReturn); err != nil {
		return err
	}

	var q bson.Document
	if err := q.ReadFrom(bufr); err != nil {
		return err
	}
	query.Query = types.MustConvertDocument(&q)

	if _, err := bufr.Peek(1); err == nil {
		var r bson.Document
		if err := r.ReadFrom(bufr); err != nil {
			return err
		}

		tr := types.MustConvertDocument(&r)
		query.ReturnFieldsSelector = &tr
	}

	return nil
}

// UnmarshalBinary reads an OpQuery from a byte array.
func (query *OpQuery) UnmarshalBinary(b []byte) error {
	br := bytes.NewReader(b)
	bufr := bufio.NewReader(br)

	if err := query.readFrom(bufr); err != nil {
		return lazyerrors.Errorf("wire.OpQuery.UnmarshalBinary: %w", err)
	}

	if _, err := bufr.Peek(1); err != io.EOF {
		return lazyerrors.Errorf("unexpected end of the OpQuery: %v", err)
	}

	return nil
}

// MarshalBinary writes an OpQuery to a byte array.
func (query *OpQuery) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	if err := binary.Write(bufw, binary.LittleEndian, query.Flags); err != nil {
		return nil, err
	}

	if err := bson.CString(query.FullCollectionName).WriteTo(bufw); err != nil {
		return nil, err
	}

	if err := binary.Write(bufw, binary.LittleEndian, query.NumberToSkip); err != nil {
		return nil, err
	}
	if err := binary.Write(bufw, binary.LittleEndian, query.NumberToReturn); err != nil {
		return nil, err
	}

	if err := bson.MustConvertDocument(query.Query).WriteTo(bufw); err != nil {
		return nil, err
	}

	if query.ReturnFieldsSelector != nil {
		if err := bson.MustConvertDocument(query.ReturnFieldsSelector).WriteTo(bufw); err != nil {
			return nil, err
		}
	}

	if err := bufw.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MarshalBinary writes an OpQuery in JSON format to byte array.
func (query *OpQuery) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"Flags":              query.Flags,
		"FullCollectionName": query.FullCollectionName,
		"NumberToSkip":       query.NumberToSkip,
		"NumberToReturn":     query.NumberToReturn,
		"Query":              bson.MustConvertDocument(query.Query),
	}
	if query.ReturnFieldsSelector != nil {
		m["ReturnFieldsSelector"] = bson.MustConvertDocument(query.ReturnFieldsSelector)
	}

	return json.Marshal(m)
}

// check interfaces
var (
	_ MsgBody = (*OpQuery)(nil)
)
