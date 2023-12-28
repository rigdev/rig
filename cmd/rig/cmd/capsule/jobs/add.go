package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/google/shlex"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"gopkg.in/yaml.v3"
)

func (c *Cmd) add(ctx context.Context, _ *cobra.Command, _ []string) error {
	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil {
		return err
	}

	allJobs := rollout.GetConfig().GetCronJobs()
	var job *capsule.CronJob
	if path == "" {
		job, err = c.promptCronJob(allJobs)
	} else {
		job, err = c.cronJobFromPath(path)
	}
	if err != nil {
		return err
	}

	for _, jj := range allJobs {
		if jj.GetJobName() == job.GetJobName() {
			return fmt.Errorf("cronjob with name %q already exists", jj.GetJobName())
		}
	}

	allJobs = append(allJobs, job)

	if err := capsule_cmd.Deploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{{
			Field: &capsule.Change_CronJobs_{
				CronJobs: &capsule.Change_CronJobs{
					Jobs: allJobs,
				},
			},
		}},
		ProjectId:     c.Cfg.GetProject(),
		EnvironmentId: base.Flags.Environment,
	}), false); err != nil {
		return err
	}

	fmt.Println("Cronjobs successfully configured!")

	return nil
}

func (c *Cmd) cronJobFromPath(path string) (*capsule.CronJob, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw interface{}
	if err := yaml.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}

	if bytes, err = json.Marshal(raw); err != nil {
		return nil, err
	}

	var job capsule.CronJob
	if err := protojson.Unmarshal(bytes, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Cmd) promptCronJob(existingJobs []*capsule.CronJob) (*capsule.CronJob, error) {
	job := &capsule.CronJob{}

	var existingNames []string
	for _, j := range existingJobs {
		existingNames = append(existingNames, j.GetJobName())
	}
	name, err := common.PromptInput("Cronjob name:", common.ValidateAndOpt(
		common.ValidateSystemName,
		common.ValidateLength(3, v1alpha2.MaxAllowedCronJobName(capsule_cmd.CapsuleID)),
		common.ValidateUnique(existingNames),
	))
	if err != nil {
		return nil, err
	}
	job.JobName = name

	idx, _, err := common.PromptSelect("Type of job:", []string{"URL", "Command"})
	if err != nil {
		return nil, err
	}
	switch idx {
	case 0:
		url, err := promptURL()
		if err != nil {
			return nil, err
		}
		job.JobType = url
	case 1:
		cmd, err := promptCommand()
		if err != nil {
			return nil, err
		}
		job.JobType = cmd
	}

	cronExp, err := common.PromptInput("Cron schedule:", common.ValidateCronExpressionOpt)
	if err != nil {
		return nil, err
	}
	job.Schedule = cronExp

	s, err := common.PromptInput("Max Retries (defaults to 6)", common.ValidateAllowEmptyOpt(common.ValidateUInt))
	if err != nil {
		return nil, err
	}
	if s != "" {
		retries, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		job.MaxRetries = int32(retries)
	}

	ds, err := common.PromptInput("Timeout Duration (optional)", common.ValidateAllowEmptyOpt(common.ValidateDuration))
	if err != nil {
		return nil, err
	}
	if ds != "" {
		d, err := time.ParseDuration(ds)
		if err != nil {
			return nil, err
		}
		job.Timeout = durationpb.New(d)
	}

	return job, nil
}

func promptURL() (*capsule.CronJob_Url, error) {
	url := &capsule.CronJob_Url{
		Url: &capsule.JobURL{
			QueryParameters: map[string]string{},
		},
	}

	s, err := common.PromptInput("Port:", common.ValidatePortOpt)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	url.Url.Port = uint64(port)

	path, err := common.PromptInput("Path:", common.ValidateURLPathOpt)
	if err != nil {
		return nil, err
	}
	url.Url.Path = path

	// TODO Add query parameters

	return url, nil
}

func promptCommand() (*capsule.CronJob_Command, error) {
	cmd := &capsule.CronJob_Command{
		Command: &capsule.JobCommand{
			Command: "",
			Args:    []string{},
		},
	}

	s, err := common.PromptInput("Command:")
	if err != nil {
		return nil, err
	}

	// Handle errors in prompt?
	splits, err := shlex.Split(s)
	if err != nil {
		return nil, err
	}
	if len(splits) > 0 {
		cmd.Command.Command = splits[0]
		cmd.Command.Args = splits[1:]
	}

	return cmd, nil
}

func (c *Cmd) delete(ctx context.Context, _ *cobra.Command, args []string) error {
	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil {
		return err
	}

	var job string
	if len(args) == 0 {
		var jobNames []string
		for _, j := range rollout.GetConfig().GetCronJobs() {
			jobNames = append(jobNames, j.GetJobName())
		}
		if len(jobNames) == 0 {
			fmt.Println("Capsule has no jobs")
			return nil
		}
		_, job, err = common.PromptSelect("Job to delete:", jobNames, common.SelectEnableFilterOpt)
		if err != nil {
			return err
		}
	} else {
		job = args[0]
	}

	found := false
	jobs := rollout.GetConfig().GetCronJobs()
	for idx, j := range rollout.GetConfig().GetCronJobs() {
		if j.GetJobName() != job {
			continue
		}
		found = true
		jobs = append(jobs[:idx], jobs[idx+1:]...)
	}

	if !found {
		return fmt.Errorf("no job with name %s", job)
	}

	if err := capsule_cmd.Deploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{{
			Field: &capsule.Change_CronJobs_{
				CronJobs: &capsule.Change_CronJobs{
					Jobs: jobs,
				},
			},
		}},
	}), false); err != nil {
		return err
	}

	return nil
}
