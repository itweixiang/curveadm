/*
 *  Copyright (c) 2021 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2022-06-23
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"fmt"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	FORMAT_MOUNT_OPTION = "type=bind,source=%s,target=%s,bind-propagation=rshared"
)

type (
	step2WaitDone struct{}

	step2CreateNBDDevice struct {
		execOptions module.ExecOptions
	}
)

func (s *step2WaitDone) Execute(ctx *context.Context) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (s *step2CreateNBDDevice) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().ModProbe("nbd", "nbds_max=64")
	_, err := cmd.Execute(s.execOptions)
	return err
}

func getAttchMount(kind, mountPoint string) string {
	var mount string
	if kind == topology.KIND_CURVEBS {
		return mount
	}
	// FIXME: use project layout to replace "/curvefs/client/mnt" path
	return fmt.Sprintf(FORMAT_MOUNT_OPTION, mountPoint, "/curvefs/client/mnt")
}

func getMountVolumes(kind string) []step.Volume {
	volumes := []step.Volume{}
	if kind == topology.KIND_CURVEFS {
		return volumes
	}

	return []step.Volume{
		{HostPath: "/dev", ContainerPath: "/dev"},
		{HostPath: "/lib/modules", ContainerPath: "/lib/modules"},
	}
}

func NewRunPlaygroundTask(curveadm *cli.CurveAdm, pc *configure.PlaygroundConfig) (*task.Task, error) {
	kind := pc.GetKind()
	name := pc.GetName()
	containerImage := pc.GetContainIamge()
	mountPoint := pc.GetMointpoint()

	// new task
	subname := fmt.Sprintf("kind=%s name=%s image=%s", kind, name, containerImage)
	t := task.NewTask("Run Playground", subname, nil)
	var containerId string

	// add step to task
	options := curveadm.ExecOptions()
	options.ExecInLocal = true

	t.AddStep(&step2CreateNBDDevice{
		execOptions: options,
	})
	t.AddStep(&step.PullImage{
		Image:       containerImage,
		ExecOptions: options,
	})
	t.AddStep(&step.CreateContainer{
		Image:             containerImage,
		Envs:              []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Name:              name, // playground-curvebs-1656035415
		Network:           "bridge",
		Mount:             getAttchMount(kind, mountPoint),
		Volumes:           getMountVolumes(kind),
		Devices:           []string{"/dev/fuse"},
		SecurityOptions:   []string{"apparmor:unconfined"},
		LinuxCapabilities: []string{"SYS_ADMIN"},
		Ulimits:           []string{"core=-1"},
		Privileged:        true,
		Out:               &containerId,
		ExecOptions:       options,
	})
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		ExecOptions: options,
	})
	t.AddStep(&step2WaitDone{})

	return t, nil
}
