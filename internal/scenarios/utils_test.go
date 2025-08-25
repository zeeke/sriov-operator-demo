package scenarios

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()

	sriovv1.InitNicIDMapFromList([]string{
		"8086 158a 154c",
		"15b3 1013 1014",
		"15b3 1015 1016",
		"15b3 1017 1018",
		"15b3 1019 101a",
		"15b3 101b 101c",
		"15b3 101d 101e",
		"15b3 101f 101e",
		"15b3 1021 101e",
	})

	os.Exit(m.Run())

}

func assertGoldenFile(t *testing.T, objects []runtime.Object) {

	goldenFile := filepath.Join("testdata", t.Name()+".yaml")

	// Convert objects to YAML for golden file comparison
	actualYAML, err := objectsToYAML(objects)
	assert.NoError(t, err)

	// Create testdata directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(goldenFile), 0755)
	assert.NoError(t, err)

	if *update {
		err = os.WriteFile(goldenFile, []byte(actualYAML), 0644)
		assert.NoError(t, err)
		t.Logf("Update golden file: %s", goldenFile)
		return
	}

	if _, err := os.Stat(goldenFile); os.IsNotExist(err) {
		err = os.WriteFile(goldenFile, []byte(actualYAML), 0644)
		assert.NoError(t, err)
		t.Logf("Created golden file: %s", goldenFile)
		return
	}

	// Read golden file
	expectedYAML, err := os.ReadFile(goldenFile)
	assert.NoError(t, err)

	assert.Equal(t, string(expectedYAML), actualYAML, "Generated YAML doesn't match golden file. Run test with -update flag to update golden file")
}

// objectsToYAML converts a slice of runtime.Object to YAML string
func objectsToYAML(objects []runtime.Object) (string, error) {
	scheme := runtime.NewScheme()
	err := sriovv1.AddToScheme(scheme)
	if err != nil {
		return "", err
	}

	err = corev1.AddToScheme(scheme)
	if err != nil {
		return "", err
	}

	err = appsv1.AddToScheme(scheme)
	if err != nil {
		return "", err
	}

	yamlPrinter := printers.NewTypeSetter(scheme).
		ToPrinter(&printers.YAMLPrinter{})

	var b strings.Builder

	for _, x := range objects {
		err := yamlPrinter.PrintObj(x, &b)
		if err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

func mockEnv(t *testing.T, key, value string) func() {
	err := os.Setenv(key, value)
	assert.NoError(t, err)

	return func() {
		err := os.Unsetenv(key)
		assert.NoError(t, err)
	}
}
