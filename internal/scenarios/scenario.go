package scenarios

import (
	"fmt"
	"os"

	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

type scenarioFactory func() ([]runtime.Object, error)

var Index map[string]scenarioFactory = map[string]scenarioFactory{}

func Dump(factory scenarioFactory) error {
	resources, err := factory()
	if err != nil {
		return fmt.Errorf("can't evaluate scenario: %w", err)
	}

	scheme := runtime.NewScheme()
	err = sriovv1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	err = corev1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	err = appsv1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	err = nadv1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	err = rbacv1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	yamlPrinter := printers.NewTypeSetter(scheme).
		ToPrinter(&printers.YAMLPrinter{})

	for _, x := range resources {
		err := yamlPrinter.PrintObj(x, os.Stdout)
		if err != nil {
			return fmt.Errorf("can't serialize object [%#v]: %w", x, err)
		}
	}

	return nil
}
