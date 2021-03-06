/*
   Copyright 2020 SUSE

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-sql-driver/mysql"
)

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		log.Fatal("SERVICE_NAME not set")
	}

	appEnv, err := cfenv.Current()
	if err != nil {
		log.Fatal(err)
	}
	mysqlService, err := appEnv.Services.WithName(serviceName)
	if err != nil {
		log.Fatal(err)
	}

	spec, ok := mysqlService.Credentials["uri"].(string)
	if !ok {
		log.Fatal(fmt.Errorf("URI not supplied"))
	}
	uri, err := url.Parse(spec)
	if err != nil {
		log.Fatal(err)
	}
	config := mysql.NewConfig()
	config.User = uri.User.String()
	config.Net = "tcp"
	config.Addr = uri.Hostname()
	config.DBName = strings.TrimPrefix(uri.Path, "/")
	uriStr := config.FormatDSN()
	fmt.Printf("Connecting to %q\n", uriStr)

	db, err := sql.Open("mysql", uriStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(createTableStatement); err != nil {
		log.Fatal(err)
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

const createTableStatement = `
CREATE Table mits(
	id int NOT NULL AUTO_INCREMENT,
	PRIMARY KEY (id)
);
`
