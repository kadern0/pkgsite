// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package frontend

import (
	"context"
	"errors"

	"golang.org/x/pkgsite/internal"
	"golang.org/x/pkgsite/internal/derrors"
	"golang.org/x/pkgsite/internal/experiment"
	"golang.org/x/pkgsite/internal/log"
)

// LatestVersion returns the latest version of the package or module.
// The linkable form of the version is returned.
// It returns the empty string on error.
// It is intended to be used as an argument to middleware.LatestVersion.
func (s *Server) LatestVersion(ctx context.Context, packagePath, modulePath, pageType string) string {
	v, err := s.latestVersion(ctx, packagePath, modulePath, pageType)
	if err != nil {
		// We get NotFound errors from directories; they clutter the log.
		if !errors.Is(err, derrors.NotFound) {
			log.Errorf(ctx, "GetLatestVersion: %v", err)
		}
		return ""
	}
	return v
}

// TODO(https://github.com/golang/go/issues/40107): this is currently tested in server_test.go, but
// we should add tests for this function.
func (s *Server) latestVersion(ctx context.Context, packagePath, modulePath, pageType string) (_ string, err error) {
	defer derrors.Wrap(&err, "latestVersion(ctx, %q, %q)", modulePath, packagePath)
	if experiment.IsActive(ctx, internal.ExperimentUsePathInfo) {
		fullPath := packagePath
		if pageType == "mod" {
			fullPath = modulePath
		}
		modulePath, version, _, err := s.ds.GetPathInfo(ctx, fullPath, modulePath, internal.LatestVersion)
		if err != nil {
			return "", err
		}
		return linkVersion(version, modulePath), nil
	}

	var mi *internal.LegacyModuleInfo
	switch pageType {
	case "mod":
		mi, err = s.ds.LegacyGetModuleInfo(ctx, modulePath, internal.LatestVersion)
		if err != nil {
			return "", err
		}
	case "pkg":
		pkg, err := s.ds.LegacyGetPackage(ctx, packagePath, modulePath, internal.LatestVersion)
		if err != nil {
			return "", err
		}
		mi = &pkg.LegacyModuleInfo
	default:
		// For directories we don't have a well-defined latest version.
		return "", nil
	}
	return linkVersion(mi.Version, modulePath), nil
}
