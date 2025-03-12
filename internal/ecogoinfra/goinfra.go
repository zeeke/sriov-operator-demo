package ecogoinfra

import (
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

var Stub *clients.Settings

func init() {
	Stub = &clients.Settings{
		AppsV1Interface: &MockAppsV1Interface{},
	}
}
