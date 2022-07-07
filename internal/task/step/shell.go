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
 * Created Date: 2021-12-13
 * Author: Jingli Chen (Wine93)
 */

package step

import (
	"strings"

	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	ERR_NOT_MOUNTED          = "not mounted"
	ERR_MOUNTPOINT_NOT_FOUND = "mountpoint not found"
	ERROR_DEVICE_BUSY        = "Device or resource busy"
)

type (
	CreateDirectory struct {
		Paths         []string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	RemoveFile struct {
		Files         []string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	CreateFilesystem struct {
		Device        string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	MountFilesystem struct {
		Source        string
		Directory     string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	UmountFilesystem struct {
		Directorys     []string
		IgnoreUmounted bool
		IgnoreNotFound bool
		ExecWithSudo   bool
		ExecInLocal    bool
		ExecSudoAlias  string
	}

	Fuser struct {
		Names         []string
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	// see also: https://linuxize.com/post/how-to-check-disk-space-in-linux-using-the-df-command/#output-format
	ShowDiskFree struct {
		Files         []string
		Format        string
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	ListBlockDevice struct {
		Device        []string
		Format        string
		NoHeadings    bool
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	ShellCommand struct {
		Command       string
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}
)

func (s *CreateDirectory) Execute(ctx *context.Context) error {
	for _, path := range s.Paths {
		if len(path) == 0 {
			continue
		}

		cmd := ctx.Module().Shell().Mkdir(path)
		cmd.AddOption("--parents") // no error if existing, make parent directories as needed
		_, err := cmd.Execute(module.ExecOption{
			ExecWithSudo:  s.ExecWithSudo,
			ExecInLocal:   s.ExecInLocal,
			ExecSudoAlias: s.ExecSudoAlias,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RemoveFile) Execute(ctx *context.Context) error {
	for _, file := range s.Files {
		if len(file) == 0 {
			continue
		}

		cmd := ctx.Module().Shell().Remove(file)
		cmd.AddOption("--force")     // ignore nonexistent files and arguments, never prompt
		cmd.AddOption("--recursive") // remove directories and their contents recursively
		out, err := cmd.Execute(module.ExecOption{
			ExecWithSudo:  s.ExecWithSudo,
			ExecInLocal:   s.ExecInLocal,
			ExecSudoAlias: s.ExecSudoAlias,
		})
		// device busy: maybe directory is mount point
		out = strings.TrimSuffix(out, "\n")
		if err != nil && !strings.HasSuffix(out, ERROR_DEVICE_BUSY) {
			return err
		}
	}
	return nil
}

func (s *CreateFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Mkfs(s.Device)
	// force mke2fs to create a filesystem, even if the specified device is not a partition
	// on a block special device, or if other parameters do not make sense
	cmd.AddOption("-F")
	_, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *MountFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Mount(s.Source, s.Directory)
	_, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *UmountFilesystem) Execute(ctx *context.Context) error {
	for _, directory := range s.Directorys {
		if len(directory) == 0 {
			continue
		}

		cmd := ctx.Module().Shell().Umount(directory)
		out, err := cmd.Execute(module.ExecOption{
			ExecWithSudo:  s.ExecWithSudo,
			ExecInLocal:   s.ExecInLocal,
			ExecSudoAlias: s.ExecSudoAlias,
		})

		out = strings.TrimSuffix(out, "\n")
		if (s.IgnoreUmounted && strings.Contains(out, ERR_NOT_MOUNTED)) ||
			(s.IgnoreNotFound && strings.Contains(out, ERR_MOUNTPOINT_NOT_FOUND)) {
			continue
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (s *Fuser) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Fuser(s.Names...)
	out, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	if err != nil {
		return err
	}

	*s.Out = strings.TrimSuffix(out, "\n")
	return nil
}

func (s *ShowDiskFree) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().DiskFree(s.Files...)
	if len(s.Format) > 0 {
		cmd.AddOption("--output=%s", s.Format)
	}

	out, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	if err != nil {
		return err
	}

	*s.Out = strings.TrimSuffix(out, "\n")
	return nil
}

func (s *ListBlockDevice) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().LsBlk(s.Device...)
	if len(s.Format) > 0 {
		cmd.AddOption("--output=%s", s.Format)
	}
	if s.NoHeadings {
		cmd.AddOption("--noheadings")
	}

	out, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	if err != nil {
		return err
	}

	*s.Out = strings.TrimSuffix(out, "\n")
	return nil
}

func (s *ShellCommand) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Command(s.Command)

	out, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	if err != nil {
		return err
	}

	if s.Out != nil {
		*s.Out = strings.TrimSuffix(out, "\n")
	}
	return nil
}
