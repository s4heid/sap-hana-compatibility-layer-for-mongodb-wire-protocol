// SPDX-FileCopyrightText: 2021 FerretDB Inc.
//
// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company
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

package handlers

import (
	"context"
	"errors"

	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/handlers/common"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/types"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/lazyerrors"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/wire"
)

type command struct {
	name           string
	help           string
	handler        func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
	storageHandler func(common.Storage, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

// Commented out commands are not supported yet
var commands = map[string]command{
	"buildInfo": {
		// db.runCommand({buildInfo: 1})
		name:    "buildInfo",
		help:    "Returns a summary of the build information.",
		handler: (*Handler).MsgBuildInfo,
	},
	"usersInfo": {
		name:    "usersInfo",
		help:    "Returns user USERNAME. Is used as a workaround to allow use of some GUIs",
		handler: (*Handler).MsgUsersInfo,
	},
	"rolesInfo": {
		name:    "rolesInfo",
		help:    "Return role readWrite. Is used as a workaround to allow use of some GUIs",
		handler: (*Handler).MsgRolesInfo,
	},
	"getlasterror": {
		name:    "getlasterror",
		help:    "Does not return last error. Is used as a workaround to allow use of some GUIs.",
		handler: (*Handler).MsgGetLastError,
	},
	"getLastError": {
		name:    "getLastError",
		help:    "Does not return last error. Is used as a workaround to allow use of some GUIs.",
		handler: (*Handler).MsgGetLastError,
	},
	"connectionStatus": {
		name:    "connectionStatus",
		help:    "checks connection",
		handler: (*Handler).MsgConnectionStatus,
	},
	// "collstats": {
	// 	// This command implements the following database methods:
	// 	// 	- db.collection.stats()
	// 	// 	- db.collection.dataSize()
	// 	name:    "collStats",
	// 	help:    "Storage data for a collection. Still needs to be implemented",
	// 	handler: (*Handler).MsgCollStats,
	// },
	// "createindexes": {
	// 	name:           "createIndexes",
	// 	help:           "Creates indexes on a collection. Still needs to be implemented.",
	// 	storageHandler: (common.Storage).MsgCreateIndexes,
	// },
	"create": {
		// db.createCollection()
		name:    "create",
		help:    "Creates the collection.",
		handler: (*Handler).MsgCreate,
	},
	// "datasize": {
	// 	// db.runCommand({dataSize: "database.collection"})
	// 	name:    "dataSize",
	// 	help:    "Returns the size of the collection in bytes.",
	// 	handler: (*Handler).MsgDataSize,
	// },
	"dbStats": {
		// db.runCommand({dbStats: 1})
		name:    "dbStats",
		help:    "Returns the statistics of the database.",
		handler: (*Handler).MsgDBStats,
	},
	"drop": {
		// db.collection.drop()
		name:    "drop",
		help:    "Drops the collection.",
		handler: (*Handler).MsgDrop,
	},
	"dropDatabase": {
		// db.dropDatabase()
		name:    "dropDatabase",
		help:    "Deletes the database.",
		handler: (*Handler).MsgDropDatabase,
	},
	// "getcmdlineopts": {
	// 	// db.adminCommand( { getCmdLineOpts: 1  } )
	// 	name:    "getCmdLineOpts",
	// 	help:    "Returns a summary of all runtime and configuration options.",
	// 	handler: (*Handler).MsgGetCmdLineOpts,
	// },
	"getLog": {
		// db.adminCommand( { getLog: "startupWarnings" } )
		name:    "getLog",
		help:    "Returns the most recent logged events from memory.",
		handler: (*Handler).MsgGetLog,
	},
	// "getparameter": {
	// 	// db.adminCommand( { getParameter : 1} )db
	// 	name:    "getParameter",
	// 	help:    "Returns the value of the parameter.",
	// 	handler: (*Handler).MsgGetParameter,
	// },
	"hostInfo": {
		// db.hostInfo()
		name:    "hostInfo",
		help:    "Returns a summary of the system information.",
		handler: (*Handler).MsgHostInfo,
	},
	"isMaster": {
		// db.isMaster()
		name:    "isMaster",
		help:    "Returns the role of the SAP HANA compatibility layer for MongoDB Wire Protocol instance.",
		handler: (*Handler).MsgHello,
	},
	"hello": {
		// db.hello()
		name:    "hello",
		help:    "Returns the role of the SAP HANA compatibility layer for MongoDB Wire Protocol instance.",
		handler: (*Handler).MsgHello,
	},
	"listCollections": {
		// db.getCollectionNames() or show collections
		name:    "listCollections",
		help:    "Returns the information of the collections and views in the database.",
		handler: (*Handler).MsgListCollections,
	},
	"listDatabases": {
		// db.adminCommand( { listDatabases: 1 } ) or show dbs
		name:    "listDatabases",
		help:    "Returns a summary of all the databases.",
		handler: (*Handler).MsgListDatabases,
	},
	"listCommands": {
		// db.listCommands()
		name: "listCommands",
		help: "Returns information about the currently supported commands.",
	},
	"ping": {
		// db.runCommand( { ping: 1 }  )
		name:    "ping",
		help:    "Returns a pong response. Used for testing purposes.",
		handler: (*Handler).MsgPing,
	},
	"whatsmyuri": {
		//  db.runCommand( { whatsmyuri: 1 } )
		name:    "whatsmyuri",
		help:    "An internal command.",
		handler: (*Handler).MsgWhatsMyURI,
	},
	"authenticate": {
		// So far only used for the authenticate required by MongoDB drivers when using tls
		// At the moment it just sends ok back to MongoDB
		name:    "authenticate",
		help:    "a method for authentication",
		handler: (*Handler).MsgAuthenticate,
	},
	// "serverstatus": {
	// 	// db.serverStatus()
	// 	name:    "serverStatus",
	// 	help:    "Returns an overview of the databases state.",
	// 	handler: (*Handler).MsgServerStatus,
	// },
	"delete": {
		// db.collection.deleteOne() or db.collection.deleteMany()
		name:           "delete",
		help:           "Deletes documents matched by the query.",
		storageHandler: (common.Storage).MsgDelete,
	},
	"find": {
		// db.collection.find()
		name:           "find",
		help:           "Returns documents matched by the custom query.",
		storageHandler: (common.Storage).MsgFindOrCount,
	},
	"findAndModify": {
		// db.collection.findandmodify()
		name:           "findAndModify",
		help:           "find one document, modifies it and return either the old document or the new document.",
		storageHandler: (common.Storage).MsgFindAndModify,
	},
	"count": {
		// db.collection.find().count()
		name:           "count",
		help:           "Returns the count of documents that's matched by the query.",
		storageHandler: (common.Storage).MsgFindOrCount,
	},
	"insert": {
		// db.collection.insertOne() or db.collection.deleteMany()
		name:           "insert",
		help:           "Inserts documents into the database.",
		storageHandler: (common.Storage).MsgInsert,
	},
	"update": {
		// db.collection.updateOne() or db.collection.updateMany()
		name:           "update",
		help:           "Updates documents that are matched by the query.",
		storageHandler: (common.Storage).MsgUpdate,
	},
	"debug_error": {
		// db.runCommand({debug_error: 1})
		name: "debug_error",
		help: "Used for debugging purposes.",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			return nil, errors.New("debug_error")
		},
	},
	"debug_panic": {
		// db.runCommand({debug_panic: 1})
		name: "debug_panic",
		help: "Used for debugging purposes.",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			panic("debug_panic")
		},
	},
}

// SupportedCommands returns a list of currently supported commands.
func SupportedCommands(context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg

	cmdList := types.MustMakeDocument()
	for _, command := range commands {
		cmdList.Set(command.name, types.MustMakeDocument(
			"help", command.help,
		))
	}

	err := reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"commands", cmdList,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
