package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/util"
)

type Run struct {
	*Base[*instructions.RunCommand, *backend_instruction.Run]
}

func NewRun(name stage.StageName, i *dockerfile.DockerfileStageInstruction[*instructions.RunCommand], dependencies []*config.Dependency, hasPrevStage bool, opts *stage.BaseStageOptions) *Run {
	return &Run{Base: NewBase(name, i, backend_instruction.NewRun(i.Data), dependencies, hasPrevStage, opts)}
}

func (stg *Run) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	args, err := stg.getDependencies(ctx, c, cb, prevImage, prevBuiltImage, buildContextArchive, stg)
	if err != nil {
		return "", err
	}

	network := instructions.GetNetwork(stg.instruction.Data)
	security := instructions.GetSecurity(stg.instruction.Data)
	mounts := instructions.GetMounts(stg.instruction.Data)

	args = append(args, "Instruction", stg.instruction.Data.Name())
	args = append(args, append([]string{"Command"}, stg.instruction.Data.CmdLine...)...)
	args = append(args, "PrependShell", fmt.Sprintf("%v", stg.instruction.Data.PrependShell))
	args = append(args, "Network", network)
	args = append(args, "Security", security)

	if len(mounts) > 0 {
		args = append(args, "Mounts")
		for _, mnt := range mounts {
			args = append(args, "Type", mnt.Type)
			args = append(args, "From", mnt.From)
			args = append(args, "Source", mnt.Source)
			args = append(args, "Target", mnt.Target)
			args = append(args, "ReadOnly", fmt.Sprintf("%v", mnt.ReadOnly))
			args = append(args, "CacheID", mnt.CacheID)
			args = append(args, "CacheSharing", mnt.CacheSharing)
			args = append(args, "Required", fmt.Sprintf("%v", mnt.Required))
			if mnt.Mode != nil {
				args = append(args, "Mode", fmt.Sprintf("%d", *mnt.Mode))
			}
			if mnt.UID != nil {
				args = append(args, "UID", fmt.Sprintf("%d", *mnt.UID))
			}
			if mnt.GID != nil {
				args = append(args, "GID", fmt.Sprintf("%d", *mnt.GID))
			}
		}
	}

	if stg.UsesBuildContext() {
		var paths []string
		for _, mnt := range mounts {
			if mnt.Type != instructions.MountTypeBind || mnt.Source == "" {
				continue
			}
			paths = append(paths, mnt.Source)
		}

		if len(paths) > 0 {
			if srcChecksum, err := buildContextArchive.CalculatePathsChecksum(ctx, paths); err != nil {
				return "", fmt.Errorf("unable to calculate build context paths checksum: %w", err)
			} else {
				args = append(args, "SourcesChecksum", srcChecksum)
			}
		}
	}

	return util.Sha256Hash(args...), nil
}
