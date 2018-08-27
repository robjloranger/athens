package fs

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	opentracing "github.com/opentracing/opentracing-go"
)

// Delete removes a specific version of a module.
func (v *storageImpl) Delete(ctx context.Context, module, version string) error {
	const op errors.Op = "fs.Delete"
	sp, ctx := opentracing.StartSpanFromContext(ctx, "storage.fs.Delete")
	defer sp.Finish()
	versionedPath := v.versionLocation(module, version)
	exists, err := v.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	if !exists {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}
	return v.filesystem.RemoveAll(versionedPath)
}
