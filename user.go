package buildah

import (
	"os/user"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

func getUser(rootdir, userspec string) (specs.User, error) {
	var gid64 uint64
	var gerr error = user.UnknownGroupError("error looking up group")

	spec := strings.SplitN(userspec, ":", 2)
	userspec = spec[0]
	groupspec := ""
	if userspec == "" {
		return specs.User{}, nil
	}
	if len(spec) > 1 {
		groupspec = spec[1]
	}

	uid64, uerr := strconv.ParseUint(userspec, 10, 32)
	if uerr == nil && groupspec == "" {
		// We parsed the user name as a number, and there's no group
		// component, so we need to look up the user's primary GID.
		var name string
		name, gid64, gerr = lookupGroupForUIDInContainer(rootdir, uid64)
		if gerr == nil {
			userspec = name
		}
	}
	if uerr != nil {
		uid64, gid64, uerr = lookupUserInContainer(rootdir, userspec)
		gerr = uerr
	}

	if groupspec != "" {
		gid64, gerr = strconv.ParseUint(groupspec, 10, 32)
		if gerr != nil {
			gid64, gerr = lookupGroupInContainer(rootdir, groupspec)
		}
	}

	if uerr == nil && gerr == nil {
		u := specs.User{
			UID:      uint32(uid64),
			GID:      uint32(gid64),
			Username: userspec,
		}
		return u, nil
	}

	err := errors.Wrapf(uerr, "error determining run uid")
	if uerr == nil {
		err = errors.Wrapf(gerr, "error determining run gid")
	}
	return specs.User{}, err
}
