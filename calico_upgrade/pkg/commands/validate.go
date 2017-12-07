// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/docopt/docopt-go"

	"github.com/projectcalico/calico/calico_upgrade/pkg/clients"
	"github.com/projectcalico/calico/calico_upgrade/pkg/commands/constants"
	"github.com/projectcalico/calico/calico_upgrade/pkg/migrate"
)

func Validate(args []string) {
	doc := constants.DatastoreIntro + `Usage:
  calico-upgrade validate
      [--apiconfigv3=<V3_APICONFIG>]
      [--apiconfigv1=<V1_APICONFIG>]
      [--output-dir=<OUTPUTDIR>]

Example:
  calico-upgrade validate --apiconfigv3=/path/to/v3/config --apiconfigv1=/path/to/v1/config

Options:
  -h --help                    Show this screen.
  --apiconfigv3=<V3_APICONFIG> Path to the file containing connection
                               configuration in YAML or JSON format for
                               the Calico v1 API.
                               [default: ` + constants.DefaultConfigPathV3 + `]
  --apiconfigv1=<V1_APICONFIG> Path to the file containing connection
                               configuration in YAML or JSON format for
                               the Calico v3 API.
                               [default: ` + constants.DefaultConfigPathV1 + `]
  --output-dir=<OUTPUTDIR>     Directory in which the data migration reports
                               are written to.
                               [default: ` + constants.GetDefaultOutputDir() + `]

Description:
  Validate that the Calico v1 format data can be migrated to Calico v3 format
  required by Calico v3.0+.

  This command generates the following set of reports (if it contains no data
  an individual report is not generated).

` + constants.ReportHelp
	parsedArgs, err := docopt.Parse(doc, args, true, "", false, false)
	if err != nil {
		fmt.Printf("Invalid option: 'calico-upgrade %s'. Use flag '--help' to read about a specific subcommand.\n", strings.Join(args, " "))
		os.Exit(1)
	}
	if len(parsedArgs) == 0 {
		return
	}
	cfv3 := parsedArgs["--apiconfigv3"].(string)
	cfv1 := parsedArgs["--apiconfigv1"].(string)
	output := parsedArgs["--output-dir"].(string)

	// Ensure we are able to write the output report to the designated output directory.
	ensureDirectory(output)

	// Obtain the v1 and v3 clients.
	clientv3, clientv1, err := clients.LoadClients(cfv3, cfv1)
	if err != nil {
		printFinalMessage("Failed to validate v1 to v3 conversion.\n"+
			"Error accessing the Calico API: %v", err)
		os.Exit(1)
	}

	// Ensure the migration code displays messages (this is basically indicating that it
	// is being called from the calico-upgrade script).
	migrate.DisplayStatusMessages(true)

	// Perform the data validation.  The validation result can only be OK or Fail.  The
	// Fail case may or may not have associated conversion data.
	data, res := migrate.Validate(clientv3, clientv1)
	if res == migrate.ResultOK {
		// We validated the data successfully.  Include a report.
		printFinalMessage("Successfully validated v1 to v3 conversion.\n" +
			"See reports below for details of the conversion.")
		printAndOutputReport(output, data)
	} else {
		if data == nil || !data.HasErrors() {
			// We failed to migrate the data and it is not due to conversion errors.  In this
			// case refer to previous messages.
			printFinalMessage("Failed to validate v1 to v3 conversion.\n" +
				"See previous messages for details.")
		} else {
			// We failed to migrate the data and it appears to be due to conversion errors.
			// In this case refer to the report for details.
			printFinalMessage("Failed to validate v1 to v3 conversion.\n" +
				"See reports blow for details of any conversion errors.")
			printAndOutputReport(output, data)
		}
		os.Exit(1)
	}

	if data == nil {
		printFinalMessage("Failed to validate v1 to v3 conversion.\n" +
			"See previous messages for details.")
		os.Exit(1)
	} else if data.HasErrors() {
		printFinalMessage("Failed to validate v1 to v3 conversion.\n" +
			"See reports below for details on any conversion errors.")
		printAndOutputReport(output, data)
		os.Exit(1)
	} else if res != migrate.ResultOK {
		printFinalMessage("Failed to validate v1 to v3 conversion.\n" +
			"See previous messages for details.")
		os.Exit(1)
	} else {
		printFinalMessage("Successfully validated v1 to v3 conversion.\n" +
			"See reports below for details of the conversion.")
		printAndOutputReport(output, data)
	}
}
