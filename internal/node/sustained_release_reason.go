package node

import "github.com/syntopica/worker/internal/host"

// sustainedReleaseReason is releaseReason with pressure counted only once it
// has lasted pressureSamples consecutive samples.
func sustainedReleaseReason(state host.State, runWhen string, pressed int) reason {
	why := releaseReason(state, runWhen)
	if why != memoryPressure || pressed >= pressureSamples {
		return why
	}

	state.Pressure = host.PressureNormal

	return releaseReason(state, runWhen)
}
