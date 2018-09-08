/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package containerd

import (
	"context"
	"errors"
	"syscall"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// NewTaskOpts allows the caller to set options on a new task
type NewTaskOpts func(context.Context, *Client, *TaskInfo) error

// WithRootFS allows a task to be created without a snapshot being allocated to its container
func WithRootFS(mounts []mount.Mount) NewTaskOpts {
	return func(ctx context.Context, c *Client, ti *TaskInfo) error {
		ti.RootFS = mounts
		return nil
	}
}

// WithCheckpointName sets the image name for the checkpoint
func WithCheckpointName(name string) CheckpointTaskOpts {
	return func(r *CheckpointTaskInfo) error {
		r.Name = name
		return nil
	}
}

// ProcessDeleteOpts allows the caller to set options for the deletion of a task
type ProcessDeleteOpts func(context.Context, Process) error

// WithProcessKill will forcefully kill and delete a process
func WithProcessKill(ctx context.Context, p Process) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// ignore errors to wait and kill as we are forcefully killing
	// the process and don't care about the exit status
	s, err := p.Wait(ctx)
	if err != nil {
		return err
	}
	if err := p.Kill(ctx, syscall.SIGKILL, WithKillAll); err != nil {
		if errdefs.IsFailedPrecondition(err) || errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}
	// wait for the process to fully stop before letting the rest of the deletion complete
	<-s
	return nil
}

// KillInfo contains information on how to process a Kill action
type KillInfo struct {
	// All kills all processes inside the task
	// only valid on tasks, ignored on processes
	All bool
	// ExecID is the ID of a process to kill
	ExecID string
}

// KillOpts allows options to be set for the killing of a process
type KillOpts func(context.Context, *KillInfo) error

// WithKillAll kills all processes for a task
func WithKillAll(ctx context.Context, i *KillInfo) error {
	i.All = true
	return nil
}

// WithKillExecID specifies the process ID
func WithKillExecID(execID string) KillOpts {
	return func(ctx context.Context, i *KillInfo) error {
		i.ExecID = execID
		return nil
	}
}

// WithResources sets the provided resources for task updates. Resources must be
// either a *specs.LinuxResources or a *specs.WindowsResources
func WithResources(resources interface{}) UpdateTaskOpts {
	return func(ctx context.Context, client *Client, r *UpdateTaskInfo) error {
		switch resources.(type) {
		case *specs.LinuxResources:
		case *specs.WindowsResources:
		default:
			return errors.New("WithResources requires a *specs.LinuxResources or *specs.WindowsResources")
		}

		r.Resources = resources
		return nil
	}
}
