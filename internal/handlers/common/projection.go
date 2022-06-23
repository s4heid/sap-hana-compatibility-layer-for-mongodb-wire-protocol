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

package common

import (
	"fmt"
	"strings"

	"github.com/DocStore/HANA_HWY/internal/types"
	"github.com/DocStore/HANA_HWY/internal/util/lazyerrors"
)

func Projection(projection types.Document) (sql string, exclusion bool, err error) {
	unimplementedFields := []string{
		"$",
		"$elemMatch",
		"$meta",
		"$slice",
		"$comment",
		"$rand",
	}

	if err = Unimplemented(&projection, unimplementedFields...); err != nil {
		return
	}

	projectionMap := projection.Map()
	if len(projectionMap) == 0 {
		sql = "*"
		return
	}

	inclusion, err := isProjectionInclusion(projection)
	if err != nil {
		return
	}

	if inclusion {
		sql = inclusionProjection(projection)
		return
	} else {
		exclusion = true
		sql = "*"
		return
	}
}

func isProjectionInclusion(projection types.Document) (inclusion bool, err error) {
	var exclusion bool
	for _, k := range projection.Keys() {
		if k == "_id" { // _id is a special case and can be both
			continue
		}

		var v any
		v, err = projection.Get(k)
		if err != nil {
			err = lazyerrors.Errorf("no value for %s.", k)
			return
		}
		switch v := v.(type) {
		case bool:
			if v {
				if exclusion {

					err = lazyerrors.Errorf("Cannot do inclusion on field #{k} in exclusion projection")
					return
				}
				if strings.Contains(k, ".") {
					err = lazyerrors.Errorf("Projection on nested documents is not implemented, yet.")
					return
				}
				inclusion = true
			} else {
				if inclusion {
					err = lazyerrors.Errorf("Cannot do exclusion on field #{k} in inclusion projection")
					return
				}
				exclusion = true
			}

		case int32, int64, float64:
			var equal types.CompareResult
			equal = 0
			if types.CompareScalars(v, int32(0)) == equal {
				if inclusion {
					err = lazyerrors.Errorf("Cannot do exclusion on field #{k} in inclusion projection")

					return
				}
				exclusion = true
			} else {
				if exclusion {
					err = lazyerrors.Errorf("Cannot do inclusion on field #{k} in exclusion projection")
					return
				}
				if strings.Contains(k, ".") {
					err = lazyerrors.Errorf("Projection on nested documents is not implemented, yet.")
					return
				}
				inclusion = true
			}
		default:
			err = lazyerrors.Errorf("unsupported operation %s %v (%T)", k, v, v)
			return
		}
	}
	return
}

func inclusionProjection(projection types.Document) (sql string) {

	sql = "{"
	if id, err := projection.Get("_id"); err == nil {
		switch id := id.(type) {
		case bool:
			if id {
				sql += "\"_id\": \"_id\", "
			}
		case int32, int64, float64:
			var equal types.CompareResult
			equal = 0
			if types.CompareScalars(id, int32(0)) != equal {
				sql += "\"_id\": \"_id\", "
			}
		}
	} else {
		sql += "\"_id\": \"_id\", "
	}

	for i, k := range projection.Keys() {

		if k == "_id" {
			continue
		}

		if i != 0 {
			sql += ", "
		}

		sql += "\"" + k + "\": \"" + k + "\""

	}

	sql += "}"

	return
}

func ProjectDocuments(docs *types.Array, projection types.Document, exclusion bool) (err error) {
	for i := 0; i < docs.Len(); i++ {
		doc, errGet := docs.GetPointer(i)
		if errGet != nil {
			return errGet
		}
		switch docv := (*doc).(type) {
		case types.Document:
			err = projectDocument(&docv, projection, exclusion)
			*doc = docv
		default:
			err = lazyerrors.Errorf("Array contains a type not being types.Document")
		}
		if err != nil {
			return
		}
	}
	return nil
}

func projectDocument(doc *types.Document, projection types.Document, exclusion bool) (err error) {
	projectionMap := projection.Map()

	for field := range projectionMap {
		fmt.Println(field)
		if strings.Contains(field, ".") {
			var next any = doc
			var previousS string
			var previousDoc types.Document
			var ppDoc types.Document
			var ppS string
			var projErr error
			var reverse []string
			for _, s := range strings.Split(field, ".") {
				reverse = append([]string{s}, reverse...)
				switch j := next.(type) {
				case *types.Document:
					ppDoc = previousDoc
					previousS = s
					next, projErr = j.Get(s)
					if projErr != nil {
						break
					}
				case types.Document:
					ppDoc = previousDoc
					ppS = previousS
					previousS = s
					previousDoc = j
					next, projErr = j.Get(s)

					if projErr != nil {
						break
					}
				default:

				}

			}
			if projErr != nil {
				continue
			}
			previousDoc.Remove(previousS)
			ppDoc.Set(ppS, previousDoc)

		} else {
			doc.Remove(field)
		}

	}

	// for k1 := range doc.Map() {
	// 	projectionVal, ok := projectionMap[k1]

	// 	fmt.Println(k1)
	// 	if !ok {
	// 		continue
	// 	}

	// 	switch projectionVal := projectionVal.(type) { // found in the projection
	// 	case bool: // field: bool
	// 		if !projectionVal {
	// 			if strings.Contains(k1, ".") {
	// 				fmt.Println(k1)
	// 			} else {
	// 				doc.Remove(k1)
	// 			}
	// 		}

	// 	case int32, int64, float64: // field: number
	// 		var equal types.CompareResult
	// 		equal = 0
	// 		if types.CompareScalars(projectionVal, int32(0)) == equal {
	// 			if strings.Contains(k1, ".") {
	// 				fmt.Println(k1)
	// 			} else {
	// 				doc.Remove(k1)
	// 			}

	// 		}
	// 	default:
	// 		return lazyerrors.Errorf("unsupported projection operation %s %v (%T)", k1, projectionVal, projectionVal)
	// 	}
	// }

	return nil
}
