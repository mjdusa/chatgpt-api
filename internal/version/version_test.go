package version_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mjdusa/chatgpt-api/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Setup Suite
type VersionSuite struct {
	suite.Suite
}

func Test_VersionSuite(t *testing.T) {
	suite.Run(t, &VersionSuite{})
}

func (s *VersionSuite) Test_GetVersion_unpopulated() {
	expected := fmt.Sprintf("%s version: []\n- Branch:     []\n- Build Time: []\n- Commit:     []\n- Go Version: []\n", os.Args[0])

	version.AppVersion = ""
	version.Branch = ""
	version.BuildTime = ""
	version.Commit = ""
	version.GoVersion = ""

	actual := version.GetVersion()

	assert.Equal(s.T(), expected, actual, "GetVersion() unpopulated message expected '%s', but got '%s'", expected, actual)
}

func (s *VersionSuite) Test_GetVersion_populated() {
	expected := fmt.Sprintf("%s version: [v1.2.3]\n- Branch:     [main]\n- Build Time: [01/01/1970T00:00:00.0000 GMT]\n- Commit:     [1234567890abcdef]\n- Go Version: [1.20.5]\n", os.Args[0])

	version.AppVersion = "v1.2.3"
	version.Branch = "main"
	version.BuildTime = "01/01/1970T00:00:00.0000 GMT"
	version.Commit = "1234567890abcdef"
	version.GoVersion = "1.20.5"

	actual := version.GetVersion()

	assert.Equal(s.T(), expected, actual, "GetVersion() populated message expected '%s', but got '%s'", expected, actual)
}
