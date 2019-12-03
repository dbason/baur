package baur

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/simplesurance/baur/storage"
)

// BuildStatus indicates if build for a current application version exist
type BuildStatus int

const (
	_ BuildStatus = iota
	// BuildStatusInputsUndefined inputs of the application are undefined,
	BuildStatusInputsUndefined
	// BuildStatusBuildCommandUndefined build.command of the application is undefined,
	BuildStatusBuildCommandUndefined
	// BuildStatusExist a build exist
	BuildStatusExist
	// BuildStatusPending no build exist
	BuildStatusPending
)

func (b BuildStatus) String() string {
	switch b {
	case BuildStatusInputsUndefined:
		return "Inputs Undefined"
	case BuildStatusExist:
		return "Exist"
	case BuildStatusPending:
		return "Pending"
	case BuildStatusBuildCommandUndefined:
		return "Build Command Undefined"
	default:
		panic(fmt.Sprintf("incompatible BuildStatus value: %d", b))
	}
}

// GetBuildStatus calculates the total input digest of the app and checks in the
// storage if a build for this input digest already exist.
// If the function returns BuildStatusExist the returned build pointer is valid
// otherwise it is nil.
func GetBuildStatus(storer storage.Storer, app *App, branchId string, compare string) (BuildStatus, *storage.BuildWithDuration, error) {
	var build *storage.BuildWithDuration
	var err error
	var branchToUse string
	if len(app.BuildCmd) == 0 {
		return BuildStatusBuildCommandUndefined, nil, nil
	}

	if !app.HasBuildInputs() {
		return BuildStatusInputsUndefined, nil, nil
	}

	d, digestErr := app.TotalInputDigest()
	if digestErr != nil {
		return -1, nil, errors.Wrap(digestErr, "calculating total input digest failed")
	}

	if compare != "" && branchId != compare {
		branchTest, testErr := storer.AreBuildsForBranch(app.Name, branchId)
		if testErr != nil {
			return -1, nil, errors.Wrap(testErr, "Checking branch builds failed")
		}
		if branchTest {
			branchToUse = branchId
		} else {
			branchToUse = compare
		}
	}
	if app.UseLastBuild {
		build, err = storer.GetLastBuildCompareDigest(app.Name, d.String(), branchToUse)
	} else {
		build, err = storer.GetLatestBuildByDigest(app.Name, d.String(), branchToUse)
	}
	if err != nil {
		if err == storage.ErrNotExist {
			return BuildStatusPending, nil, nil
		}
		return -1, nil, errors.Wrap(err, "fetching latest build failed")
	}

	return BuildStatusExist, build, nil
}
