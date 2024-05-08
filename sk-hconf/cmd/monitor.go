package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"skas/sk-hconf/internal/global"
	"skas/sk-hconf/pkg/texttemplate"
	"sort"
	"strings"
	"time"
)

var monitorParams struct {
	remove           bool
	mark             bool
	force            bool
	ttlAfterFinished time.Duration
	jobTemplate      string
}

func init() {
	MonitorCmd.PersistentFlags().StringVar(&monitorParams.jobTemplate, "jobTemplate", "/templates/job.tmpl", "Job template file for each node")
	MonitorCmd.PersistentFlags().BoolVar(&monitorParams.remove, "remove", false, "Remove webhook configuration")
	MonitorCmd.PersistentFlags().BoolVar(&monitorParams.force, "force", false, "Perform even if apiserver is down")
	MonitorCmd.PersistentFlags().BoolVar(&monitorParams.mark, "mark", false, "Display dot on pod state change wait. Log if false")
	// This is a last resort parameter, as the child job should be cleanup up by its parent
	MonitorCmd.PersistentFlags().DurationVar(&monitorParams.ttlAfterFinished, "ttlAfterFinished", time.Minute*30, "Wait before cleanup")

	_ = MonitorCmd.MarkPersistentFlagRequired("image")
}

var MonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor SKAS authentication webhook configuration",
	Run: func(cmd *cobra.Command, args []string) {
		global.Logger.Info("Auth webhook configuration monitor", "version", global.Version, "build", global.BuildTs, "logLevel", rootParams.logConfig.Level, "remove", monitorParams.remove)

		nodes, err := lookupApiServerNodes()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error while listing APIs server nodes: %v\n", err)
			os.Exit(2)
		}
		sort.Strings(nodes) // To have a predictable order
		global.Logger.Info("Lookup api server", "nodes", nodes)

		for idx, nodeName := range nodes {
			err := handleNodeJob(context.Background(), idx, nodeName)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error on job for '%s': %v\n", nodeName, err)
				os.Exit(2)
			}
		}
	},
}

func lookupApiServerNodes() ([]string, error) {
	pods, err := global.ClientSet.CoreV1().Pods(global.Config.ApiServerNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, 3)
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, global.Config.ApiServerPodName) {
			result = append(result, pod.Spec.NodeName)
		}
	}
	return result, nil
}

func handleNodeJob(ctx context.Context, idx int, nodeName string) error {
	job, err := buildJob(idx, nodeName)
	if err != nil {
		return err
	}
	global.Logger.Info("handle node job", "jobName", job.Name, "nodeName", job.Spec.Template.Spec.NodeName, "image", job.Spec.Template.Spec.Containers[0].Image)
	_, err = global.ClientSet.BatchV1().Jobs(global.Config.SkasNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	// And loop until job end
	global.Logger.Info("Wait for child job to end", "nodeName", nodeName, "idx", idx)
	limit := time.Now().Add(global.Config.TimeoutApiServer)
	for {
		time.Sleep(time.Second)
		job2, err := global.ClientSet.BatchV1().Jobs(global.Config.SkasNamespace).Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			// Api server unreachable is a normal case on a 1 node control plane)
			if monitorParams.mark {
				fmt.Printf(":")
			}
			if time.Now().After(limit) {
				if monitorParams.mark {
					fmt.Printf("\n")
				}
				return fmt.Errorf("timeout on apiserver up expired. Last error: %v", err)
			}
		} else {
			ended, st := isJobFinished(job2)
			if ended {
				if st == batchv1.JobFailed {
					if monitorParams.mark {
						fmt.Printf("\n")
					}
					return fmt.Errorf("child job#%d failed (node:%s)", idx, nodeName)
				}
				if monitorParams.mark {
					fmt.Printf("\n")
				}
				global.Logger.Info("child job OK", "nodeName", nodeName, "idx", idx)
				return nil
			} else {
				if monitorParams.mark {
					fmt.Printf(".")
				}
			}
		}
	}
}

/*
We consider a job "finished" if it has a "Complete" or "Failed" condition marked as true.
Status conditions allow us to add extensible status information to our objects that other
humans and controllers can examine to check things like completion and health.
*/
func isJobFinished(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

func buildOwnerReference() map[string]interface{} {
	myPodName := os.Getenv("MY_POD_NAME")
	myPodUid := os.Getenv("MY_POD_UID")
	if myPodName != "" && myPodUid != "" {
		oref := make(map[string]interface{})
		oref["name"] = myPodName
		oref["uid"] = myPodUid
		global.Logger.Info("setting ownerReferences", "podName", myPodName, "uid", myPodUid)
		return oref
	} else {
		global.Logger.Info("Unable to set ownerReferences. Missing MY_POD_NAME and/or MY_POD_UID environment variables")
		return nil
	}
}

// https://github.com/kubernetes/client-go/issues/193

func buildJob(idx int, nodeName string) (*batchv1.Job, error) {

	model := map[string]interface{}{
		"Config": global.Config,
		"Values": map[string]interface{}{
			"idx":                     idx,
			"nodeName":                nodeName,
			"ownerReferences":         buildOwnerReference(),
			"ttlSecondsAfterFinished": monitorParams.ttlAfterFinished.Seconds(),
			"mark":                    monitorParams.mark,
			"remove":                  monitorParams.remove,
			"force":                   monitorParams.force,
			"log": map[string]interface{}{
				"level": rootParams.logConfig.Level,
				"mode":  rootParams.logConfig.Mode,
			},
		},
	}

	result, err := texttemplate.NewAndRenderToTextFromFile(monitorParams.jobTemplate, model)
	if err != nil {
		return nil, err
	}
	job := &batchv1.Job{}
	err = yaml.NewYAMLToJSONDecoder(strings.NewReader(result)).Decode(job)
	if err != nil {
		return nil, err
	}
	return job, nil
}
