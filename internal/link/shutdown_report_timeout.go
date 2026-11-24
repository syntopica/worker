package link

import "time"

// shutdownReportTimeout leaves time to reap runners inside the shared
// launchd shutdown budget if the coordinator is stalled.
const shutdownReportTimeout = time.Second
