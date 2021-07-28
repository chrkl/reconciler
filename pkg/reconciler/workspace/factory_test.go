package workspace

import (
	"github.com/kyma-incubator/reconciler/pkg/test"
	"path/filepath"
	"testing"

	file "github.com/kyma-incubator/reconciler/pkg/files"
	"github.com/stretchr/testify/require"
)

const version = "1.20.0"

func TestWorkspaceFactory(t *testing.T) {
	t.Run("Test validation", func(t *testing.T) {
		wsf1 := Factory{
			Debug: true,
		}
		require.NoError(t, wsf1.validate(version))
		require.Equal(t, filepath.Join(wsf1.defaultStorageDir(), version), wsf1.workspaceDir)
		require.Equal(t, defaultRepositoryURL, wsf1.RepositoryURL)

		wsf2 := Factory{
			Debug:      true,
			StorageDir: "/tmp",
		}
		require.NoError(t, wsf2.validate(version))
		require.Equal(t, filepath.Join("/tmp", version), wsf2.workspaceDir)
		require.Equal(t, defaultRepositoryURL, wsf1.RepositoryURL)
	})

	t.Run("Clone and delete workspace", func(t *testing.T) {
		if !test.RunExpensiveTests() {
			//this test case clones the Kyma repo can take up to 60 sec (depending on the bandwidth) and generates bigger amount of traffic
			return
		}

		workspaceDir := filepath.Join(".", "test", version)
		wsf := &Factory{
			StorageDir: "./test",
		}

		//cleanup at the beginning (if test was interrupted before)
		testDelete(t, wsf)
		//cleanup at the end (if test finishes regularly)
		defer testDelete(t, wsf)

		ws, err := wsf.Get(version)
		require.NoError(t, err)

		require.Equal(t, filepath.Join(workspaceDir, componentFile), ws.ComponentFile)
		require.True(t, file.Exists(ws.ComponentFile))
		require.Equal(t, filepath.Join(workspaceDir, resDir), ws.ResourceDir)
		require.True(t, file.DirExists(ws.ResourceDir))
		require.Equal(t, filepath.Join(workspaceDir, instResDir), ws.InstallationResourceDir)
		require.True(t, file.DirExists(ws.InstallationResourceDir))
	})

}

func testDelete(t *testing.T, wsf *Factory) {
	require.NoError(t, wsf.Delete(version))
}