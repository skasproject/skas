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
	timeout          time.Duration
	mark             bool
	force            bool
	image            string
	ttlAfterFinished time.Duration
}

func init() {
	MonitorCmd.PersistentFlags().BoolVar(&monitorParams.remove, "remove", false, "Remove webhook configuration")
	MonitorCmd.PersistentFlags().BoolVar(&monitorParams.force, "force", false, "Perform even if apiserver is down")
	MonitorCmd.PersistentFlags().DurationVar(&monitorParams.timeout, "timeout", time.Second*240, "Timeout on API server down or up")
	MonitorCmd.PersistentFlags().BoolVar(&monitorParams.mark, "mark", false, "Display dot on pod state change wait. Log if false")
	MonitorCmd.PersistentFlags().StringVar(&monitorParams.image, "image", "", "container image for patch")
	// This is a last resort parameter, as the child job should be cleanup up by its parent
	MonitorCmd.PersistentFlags().DurationVar(&monitorParams.ttlAfterFinished, "ttlAfterFinished", time.Minute*10, "Wait before cleanup")

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
	fmt.Printf("name:%s    nodeName:%s    image:%s\n", job.Name, job.Spec.Template.Spec.NodeName, job.Spec.Template.Spec.Containers[0].Image)
	_, err = global.ClientSet.BatchV1().Jobs(global.Config.SkasNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// And loop until job end
	global.Logger.Info("Wait for child job to end", "nodeName", nodeName, "idx", idx)
	for {
		time.Sleep(time.Second)
		job2, err := global.ClientSet.BatchV1().Jobs(global.Config.SkasNamespace).Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		ended, st := isJobFinished(job2)
		if ended {
			if st == batchv1.JobFailed {
				return fmt.Errorf("child job#%d failed (node:%s)", idx, nodeName)
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
		"Values": map[string]interface{}{
			"jobName":                 fmt.Sprintf("job-sk-hconf-%d", idx),
			"namespace":               global.Config.SkasNamespace,
			"serviceAccount":          global.Config.ServiceAccount,
			"image":                   monitorParams.image,
			"ttlSecondsAfterFinished": monitorParams.ttlAfterFinished.Seconds(),
			"nodeName":                nodeName,
			"remove":                  monitorParams.remove,
			"force":                   monitorParams.force,
			"timeout":                 monitorParams.timeout.String(),
			"mark":                    monitorParams.mark,
			"ownerReferences":         buildOwnerReference(),
			"log": map[string]interface{}{
				"level": rootParams.logConfig.Level,
				"mode":  rootParams.logConfig.Mode,
			},
		},
	}
	tmpl, err := texttemplate.New("jobTemplate", jobTemplate)
	if err != nil {
		return nil, err
	}
	result, err := tmpl.RenderToText(model)
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

const jobTemplate = `apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Values.jobName }}
  namespace: {{ .Values.namespace }}
  {{- with .Values.ownerReferences }}
  ownerReferences:
  - apiVersion: v1
    kind: Pod
    name: {{ .name }}
    uid: {{ .uid }}
    blockOwnerDeletion: true
    controller: true
  {{- end }}
spec:
  ttlSecondsAfterFinished: {{ .Values.ttlSecondsAfterFinished }}
  backoffLimit: 1
  template:
    metadata:
      labels:
        app.kubernetes.io/name: sk-hconf
        app.kubernetes.io/instance: {{ .Values.jobName }}
    spec:
      serviceAccountName: {{ .Values.serviceAccount }}
      nodeName: {{ .Values.nodeName }}
      securityContext:
        runAsUser: 0
      containers:
        - name: patch
          image: {{ .Values.image }}
          imagePullPolicy: Always
          args:
            - patch
            - --configFile
            - /config.yaml
            - --logMode
            - {{ .Values.log.mode }}
            - --logLevel
            - {{ .Values.log.level }}
            - --nodeName
            - {{ .Values.nodeName }}
            - --timeout
            - {{ .Values.timeout }}
            {{- if .Values.mark }}
            - --mark
            {{- end }}
            {{- if .Values.force }}
            - --force
            {{- end }}
            {{- if .Values.remove }}
            - --remove
            {{- end }}
          securityContext:
            allowPrivilegeEscalation: true
          volumeMounts:
            - mountPath: /etc/kubernetes
              name: kube-conf
              readOnly: false
            - mountPath: /config.yaml
              name: config
              subPath: config.yaml
      volumes:
        - name: kube-conf
          hostPath:
            path: /etc/kubernetes
            type: Directory
        - name: config
          configMap:
            name: sk-hconf
      restartPolicy: Never
`
