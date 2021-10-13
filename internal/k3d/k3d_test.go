package k3d

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if !setup() {
		fmt.Println("Setup of k3d test failed: test case for k3d can't be executed")
		return
	}
	code := m.Run()
	//shutdown()
	os.Exit(code)
}

// Place this folder at the beginning of PATH env-var to ensure this
// mock-script will be used instead of a locally installed k3d tool.
func setup() bool {
	if os.Getenv("GOPATH") == "" {
		fmt.Println("Could not inject k3d mock directory into PATH: env-var GOPATH is undefined")
		return false
	}

	currentDir, err := os.Getwd()
	fmt.Println(currentDir)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	mockDir := fmt.Sprintf("%s/mock", currentDir)

	os.Setenv("PATH", fmt.Sprintf("%s:%s", mockDir, os.Getenv("PATH")))
	return true
}

// function to verify output of k3d tool
type testFunc func(output string, err error)

func TestRunCmd(t *testing.T) {
	tests := []struct {
		cmd      []string
		verifier testFunc
	}{
		{
			cmd: []string{"cluster", "list"},
			verifier: testFunc(func(output string, err error) {
				if !strings.Contains(output, "kyma-cluster") {
					require.Fail(t, fmt.Sprintf("Expected string 'kyma-cluster' is missing in k3d output: %s", output))
				}
			}),
		},
		{
			cmd: []string{"cluster", "xyz"},
			verifier: testFunc(func(output string, err error) {
				require.NotEmpty(t, err, "Error object expected")
			}),
		},
	}

	for testID, testCase := range tests {
		output, err := RunCmd(false, 5*time.Second, testCase.cmd...)
		require.NotNilf(t, testCase.verifier, "Verifier function missing for test #'%d'", testID)
		testCase.verifier(output, err)
	}

}

func TestCheckVersion(t *testing.T) {
	err := checkVersion(false)
	require.NoError(t, err)
}

func TestCheckVersionIncompatibleMinor(t *testing.T) {
	os.Setenv("K3D_MOCK_DUMPFILE", "version_incompminor.txt")
	err := checkVersion(false)
	require.Error(t, err)
	os.Setenv("K3D_MOCK_DUMPFILE", "")
}

func TestCheckVersionIncompatibleMajor(t *testing.T) {
	os.Setenv("K3D_MOCK_DUMPFILE", "version_incompmajor.txt")
	err := checkVersion(false)
	require.Error(t, err)
	os.Setenv("K3D_MOCK_DUMPFILE", "")
}

func TestInitialize(t *testing.T) {
	err := Initialize(false)
	require.NoError(t, err)
}

func TestInitializeFailed(t *testing.T) {
	pathPrev := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/bin")

	err := Initialize(false)
	require.Error(t, err)

	os.Setenv("PATH", pathPrev)
}

func TestRegistryExists(t *testing.T) {
	exists, err := RegistryExists(false, "kyma")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestCreateRegistry(t *testing.T) {
	registryURL, err := CreateRegistry(false, 5*time.Second, "kyma")
	require.Equal(t, "kyma-registry:5000", registryURL)
	require.NoError(t, err)
}

func TestDeleteRegistry(t *testing.T) {
	err := DeleteRegistry(false, 5*time.Second, "kyma")
	require.NoError(t, err)
}

func TestStartCluster(t *testing.T) {
	k3dSettings := Settings{
		ClusterName: "kyma",
		Args:        []string{"--alsologtostderr"},
		Version:     "1.20.7",
		PortMapping: []string{"80:80@loadbalancer", "443:443@loadbalancer"},
	}
	err := StartCluster(false, 5*time.Second, 1, []string{}, []string{"k3d-kyma-registry.localhost"}, k3dSettings)
	require.NoError(t, err)
}

func TestDeleteCluster(t *testing.T) {
	err := DeleteCluster(false, 5*time.Second, "kyma")
	require.NoError(t, err)
}

func TestClusterExists(t *testing.T) {
	os.Setenv("K3D_MOCK_DUMPFILE", "cluster_list_exists.json")
	exists, err := ClusterExists(false, "kyma")
	require.NoError(t, err)
	require.True(t, exists)
	os.Setenv("K3D_MOCK_DUMPFILE", "")
}

func TestArgConstruction(t *testing.T) {
	rawPorts := []string{"8000:80@loadbalancer", "8443:443@loadbalancer"}
	res := constructArgs("-p", rawPorts)
	require.Equal(t, []string{"-p", "8000:80@loadbalancer", "-p", "8443:443@loadbalancer"}, res)
}
