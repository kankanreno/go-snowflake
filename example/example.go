package main

import (
	"fmt"
	"github.com/kankanreno/go-snowflake"
	"time"
)

func main() {
	// set starttime and machineID for the first time if you wan't to use the default value
	snowflake.SetStartTime(time.Date(1955, 1, 1, 0, 0, 0, 0, time.UTC))
	snowflake.SetMachineID(snowflake.PrivateIPToMachineID() % (snowflake.MaxMachineID + 1)) // testing, not to be used in production

	id := snowflake.ID()
	fmt.Println(id)
	// 9007199254740992 js max integer
	// 496286021178368  4-bit machine id (unset), start 2008
	// 545487571292160  6-bit machine id (unset), start 2020
	// 545489678665664  6-bit machine id (  set), start 2020
	// 2096606852595712 6-bit machine id (unset), start 2008
	// 8947297216425984 6-bit machine id (unset), start 1955

	sid := snowflake.ParseID(id)
	// SID {
	//     Sequence: 0
	//     MachineID: 0
	//     Timestamp: x
	//     ID: x
	// }
	fmt.Println(sid)
}
