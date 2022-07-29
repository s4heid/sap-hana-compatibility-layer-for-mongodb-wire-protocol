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

package bson

import (
	"bufio"

	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/fjson"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/lazyerrors"
)

// Bool represents BSON Bool data type.
type Bool bool

func (b *Bool) bsontype() {}

// ReadFrom implements bsontype interface.
func (b *Bool) ReadFrom(r *bufio.Reader) error {
	v, err := r.ReadByte()
	if err != nil {
		return lazyerrors.Errorf("bson.Bool.ReadFrom: %w", err)
	}

	switch v {
	case 0:
		*b = false
	case 1:
		*b = true
	default:
		return lazyerrors.Errorf("bson.Bool.ReadFrom: unexpected byte %#02x", v)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (b Bool) WriteTo(w *bufio.Writer) error {
	v, err := b.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Bool.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Bool.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (b Bool) MarshalBinary() ([]byte, error) {
	if b {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

// UnmarshalJSON implements bsontype interface.
func (b *Bool) UnmarshalJSON(data []byte) error {
	var bJ fjson.Bool
	if err := bJ.UnmarshalJSON(data); err != nil {
		return err
	}

	*b = Bool(bJ)
	return nil
}

// MarshalJSON implements bsontype interface.
func (b Bool) MarshalJSON() ([]byte, error) {
	return fjson.Marshal(fromBSON(&b))
}

// check interfaces
var (
	_ bsontype = (*Bool)(nil)
)
