package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// I'm not parsing default config path in home.
// I just suppose it would be convenient enough to
// pass path manually.
func getClient(args Arguments) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", args.kubeconfig)

	if err != nil {
		fmt.Printf("Error: building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error: creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}
	return clientset
}

// Generate job id to keep track of other script runs
func getJobId() string {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Error: creating job id uuid: %v\n", err)
		os.Exit(1)
	}
	return uuidObj.String()
}

// Create volume for persisting the results.
func createPersistentVolume(clientset *kubernetes.Clientset, jobId string) *v1.PersistentVolume {
	persist := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tgjobvolume-" + jobId,
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: "standard",
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("100Mi"),
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRecycle,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/tmp",
				},
			},
		},
	}
	_, err := clientset.CoreV1().PersistentVolumes().Create(context.TODO(), persist, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating PersistentVolume: %v\n", err)
		os.Exit(1)
	}

	return persist

}

func createPersistentVolumeClaim(clientSet *kubernetes.Clientset, namespace string, jobId string) *v1.PersistentVolumeClaim {

	// Define the PVC manifest
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tgjobvolumeclaim-" + jobId,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce, // This specifies that the PVC can be mounted by a single node at a time.
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("100Mi"), // The requested storage size for the PVC
				},
			},
		},
	}

	// Create the PVC
	_, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating PVC: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PersistentVolumeClaim created successfully!")
	return pvc
}

// Run Python Telegram management script
func runMainScript(clientSet *kubernetes.Clientset, jobId string, arguments Arguments, volume *v1.PersistentVolume) *v1.Pod {
	// Define xthe Pod manifest
	var volume_spec []v1.Volume
	var container_spec []v1.Container

	// resources := v1.ResourceRequirements{
	// 	Limits: v1.ResourceList{
	// 		v1.ResourceCPU:    resource.MustParse("0.5"), // CPU limit (0.5 cores)
	// 		v1.ResourceMemory: resource.MustParse("512Mi"), // Memory limit (512 MiB)
	// 	},
	// 	Requests: v1.ResourceList{
	// 		v1.ResourceCPU:    resource.MustParse("0.1"), // CPU request (0.1 cores)
	// 		v1.ResourceMemory: resource.MustParse("256Mi"), // Memory request (256 MiB)
	// 	},
	// },

	envs := []v1.EnvVar{
		{
			Name:  "TG_MANAGE_API_ID",
			Value: fmt.Sprint(arguments.api_id),
		},
		{
			Name:  "TG_MANAGE_API_HASH",
			Value: arguments.api_hash,
		},
		{
			Name:  "TG_MANAGE_BOT_TOKEN",
			Value: arguments.session,
		},
	}
	time.Sleep(1 * time.Second)

	if arguments.messageIDs != "" && arguments.chatID != 0 {
		envs = append(envs, v1.EnvVar{
			Name:  "TG_MANAGE_REMOVE_MESSAGE_IDS",
			Value: "[" + arguments.resultsFile + "]",
		})
		envs = append(envs, v1.EnvVar{
			Name:  "TG_MANAGE_REMOVE_CHAT_ID",
			Value: fmt.Sprint(arguments.chatID),
		})
	}

	if arguments.removeAll {
		envs = append(envs, v1.EnvVar{
			Name:  "TG_MANAGE_REMOVE_ALL",
			Value: "True",
		})
	} else {
		envs = append(envs, v1.EnvVar{
			Name:  "TG_MANAGE_REMOVE_ALL",
			Value: "False",
		})
	}

	if arguments.persist {
		volume_spec = nil
		container_spec = []v1.Container{
			{
				Name:    "tgjob-" + jobId,
				Image:   arguments.imageName,
				Command: []string{"python", "-m", "manage"},
				Env:     envs,
			},
		}
	} else {
		envs = append(envs, v1.EnvVar{
			Name:  "TG_MANAGE_PERSIST_PATH",
			Value: "./results/" + arguments.resultsFile,
		})
		pvc := createPersistentVolumeClaim(clientSet, arguments.namespace, jobId)

		volume_spec = []v1.Volume{
			{
				Name: "tgjobvolume-" + jobId,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
					},
				},
			},
		}
		container_spec = []v1.Container{
			{
				Name:    "tgjob-" + jobId,
				Image:   arguments.imageName,
				Command: []string{"python", "-m", "manage"},
				Env:     envs,
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "tgjobvolume-" + jobId,
						MountPath: "/usr/results",
					},
				},
			},
		}
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tgjob-" + jobId,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever, // This specifies that the Pod should not be restarted automatically
			Containers:    container_spec,
			Volumes:       volume_spec,
		},
	}
	// Create the Pod
	_, err := clientSet.CoreV1().Pods(arguments.namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating Pod: %v\n", err)
		os.Exit(1)
	}
	return pod
}

// We need to notify user about the job status
func callback_on_pod_status(pod *v1.Pod) {
	if pod.Status.Phase == v1.PodSucceeded {
		fmt.Println("Pod completed successfully.")
		os.Exit(0)
	} else if pod.Status.Phase == v1.PodFailed {
		fmt.Println("Pod failed.")
		os.Exit(1)
	}

}

// we need to wait until the job would be done
func watchPodStatus(clientset *kubernetes.Clientset, namespace string, podName string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podWatcher, err := clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: podName}))
	if err != nil {
		fmt.Printf("Error watching Pod: %v\n", err)
		os.Exit(1)
	}

	for {
		select {
		case event, isOpen := <-podWatcher.ResultChan():
			if !isOpen {
				fmt.Println("Pod watcher channel closed unexpectedly.")
				os.Exit(1)
			}
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				fmt.Println("Unexpected object type received from Pod watcher.")
				os.Exit(1)
			}
			callback_on_pod_status(pod)
		}
	}
}
