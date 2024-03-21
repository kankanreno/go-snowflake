package main

import (
	"fmt"
	"time"

	"github.com/kankanreno/go-snowflake"
)

func main() {
	// set starttime and machineID for the first time if you wan't to use the default value
	snowflake.SetStartTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	snowflake.SetMachineID(snowflake.PrivateIPToMachineID()) // testing, not to be used in production

	id := snowflake.ID()
	fmt.Println(id) // 329874157232128

	sid := snowflake.ParseID(id)
	// SID {
	//     Sequence: 0
	//     MachineID: 0
	//     Timestamp: x
	//     ID: x
	// }
	fmt.Println(sid)
}
