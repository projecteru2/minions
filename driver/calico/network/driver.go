package network

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// IFPrefix .
var IFPrefix = "cali"

func init() { // nolint
	if os.Getenv("CALICO_LIBNETWORK_IFPREFIX") != "" {
		IFPrefix = os.Getenv("CALICO_LIBNETWORK_IFPREFIX")
		log.Infof("Updated CALICO_LIBNETWORK_IFPREFIX to %s", IFPrefix)
	}
}
