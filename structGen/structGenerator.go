/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
  Unless required by applicable law or agreed to in writing,
  software distributed under the License is distributed on an
  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  KIND, either express or implied.  See the License for the
  specific language governing permissions and limitations
  under the License.
*/

//structGenerator.go
//generates 'structs' package
package structGenerator

import (
	"./../sqlParser"
	"./../structBuilder"
	"os"
	"strings"
)

var (
	username    = os.Args[1]
	password    = os.Args[2]
	environment = os.Args[3]
	db          = sqlParser.ConnectToDatabase(username, password, environment)
)

func main() {
	InitStructFiles()
}

//error checking
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func AppendToStructFiles(table string) {
	AppendToStructs(table)
	AppendToStructInterface(table)
	AppendToStructMap(table)
	AppendToStructValidMap(table)
}

//writes struct, interface, and map files to structs package
func InitStructFiles() {
	InitStructs()
	InitStructInterface()
	InitStructMap()
	InitStructValidMap()
}

func AppendToStructs(table string) {
	columnList := sqlParser.GetColumnNames(table)
	columnTypes := sqlParser.GetColumnTypes(table)

	structStr := structBuilder.MakeStructStr(table, columnList, columnTypes)

	structBuilder.AddToFile("./structs/structs.go", structStr)
}

//writes the struct file (structs.go), which has an object for each
//database table, ith each table field as a member variable
func InitStructs() {
	structStr := "package structs\n"

	tableList := sqlParser.GetTableNames()
	tableList = append(tableList, sqlParser.GetViewNames()...)

	//add a struct for each table
	for _, table := range tableList {
		columnList := sqlParser.GetColumnNames(table)
		columnTypes := sqlParser.GetColumnTypes(table)

		structStr += structBuilder.MakeStructStr(table, columnList, columnTypes)
	}

	//writes in relation to home directory
	structBuilder.WriteFile(structStr, "./structs/structs.go")
}

func AppendToStructInterface(table string) {
	//function declaration
	structInterface := "func EncodeStruct" + strings.Title(table) + "(rows *sqlx.Rows, w http.ResponseWriter) {\n" //make new array
	structInterface += "\tsa := make([]" + strings.Title(table) + ", 0)\n"
	//make new instance
	structInterface += "\tt := " + strings.Title(table) + "{}\n\n"
	//loops through all columns and translates to JSON
	structInterface += "\tfor rows.Next() {\n"
	structInterface += "\t\t rows.StructScan(&t)\n"
	structInterface += "\t\t sa = append(sa, t)\n"
	structInterface += "\t}\n\n"

	structInterface += "\twrapper := Wrapper{sa, 1.1}\n"

	structInterface += "\tenc := json.NewEncoder(w)\n"
	structInterface += "\tenc.Encode(wrapper)\n"
	structInterface += "}\n"

	structBuilder.AddToFile("./structs/structInterface.go", structInterface)
}

//writes structInterface.go, which has a function that takes in *Rows,
//makes them into an array of []TableName structs, and encodes this
//array into JSON format
func InitStructInterface() {
	//header, imports
	structInterface := "package structs\n"
	structInterface += "import (\n"
	structInterface += "\t\"github.com/jmoiron/sqlx\"\n"
	structInterface += "\t\"encoding/json\"\n"
	structInterface += "\t\"net/http\"\n"
	structInterface += ")\n"

	//makes a function for each object
	tableList := sqlParser.GetTableNames()
	tableList = append(tableList, sqlParser.GetViewNames()...)
	for _, table := range tableList {
		//function declaration
		structInterface += "func EncodeStruct" + strings.Title(table) + "(rows *sqlx.Rows, w http.ResponseWriter) {\n" //make new array
		structInterface += "\tsa := make([]" + strings.Title(table) + ", 0)\n"
		//make new instance
		structInterface += "\tt := " + strings.Title(table) + "{}\n\n"
		//loops through all columns and translates to JSON
		structInterface += "\tfor rows.Next() {\n"
		structInterface += "\t\t rows.StructScan(&t)\n"
		structInterface += "\t\t sa = append(sa, t)\n"
		structInterface += "\t}\n\n"

		structInterface += "\twrapper := Wrapper{sa, 1.1}\n"

		structInterface += "\tenc := json.NewEncoder(w)\n"
		structInterface += "\tenc.Encode(wrapper)\n"
		structInterface += "}\n"
	}

	//writes in relation to home directory
	structBuilder.WriteFile(structInterface, "./structs/structInterface.go")
}

func AppendToStructValidMap(table string) {

	structValid := "\t\"" + table + "\" : true,\n"

	structBuilder.AddToMethodInFile("./structs/structValidMap.go", structValid)
}

//writes structValidMap.go, which maps each table in the database to the boolean "true,"
//used to confirm validity of URL
func InitStructValidMap() {
	structValid := "package structs\n"

	structValid += "var ValidStruct = map[string]bool {\n"

	tableList := sqlParser.GetTableNames()
	tableList = append(tableList, sqlParser.GetViewNames()...)
	for _, table := range tableList {
		structValid += "\t\"" + table + "\" : true,\n"
	}

	structValid += "}\n"

	structBuilder.WriteFile(structValid, "./structs/structValidMap.go")
}

func AppendToStructMap(table string) {
	structMap := "\tif tableName == \"" + table + "\"{\n"
	structMap += "\t\tEncodeStruct" + strings.Title(table) + "(rows, w)\n"
	structMap += "\t}\n"

	structBuilder.AddToMethodInFile("./structs/structMap.go", structMap)
}

//writes structMap.go, which has a function that maps each tableName string to
//its respective function in structInterface.go
func InitStructMap() {
	//declaration, imports
	structMap := "package structs\n"
	structMap += "import (\n\t\"github.com/jmoiron/sqlx\"\n"
	structMap += "\t\"net/http\"\n)\n"
	structMap += "func MapTableToJson(tableName string, rows *sqlx.Rows, w http.ResponseWriter) {\n"

	//each table has a case mapping name with structInterface function
	tableList := sqlParser.GetTableNames()
	tableList = append(tableList, sqlParser.GetViewNames()...)
	for _, table := range tableList {
		structMap += "\tif tableName == \"" + table + "\"{\n"
		structMap += "\t\tEncodeStruct" + strings.Title(table) + "(rows, w)\n"
		structMap += "\t}\n"
	}

	//if invalid table, returns error
	structMap += "}\n"

	//writes in relation to home directory
	structBuilder.WriteFile(structMap, "./structs/structMap.go")
}