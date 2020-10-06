package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/kungfusoftware/externalbuilder/pkg/utils"
	"github.com/lithammer/shortuuid/v3"
	"github.com/pkg/errors"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	chaincodeSourceDir := args[0]
	//metadataDir := args[1]
	buildOutputDir := args[2]

	var err error
	// copy connection.json
	//connectionJsonPath := path.Join(chaincodeSourceDir, "connection.json")
	//connectionDestFile := path.Join(buildOutputDir, "connection.json")
	//err := utils.Copy(connectionJsonPath, connectionDestFile)
	//utils.HandleErr(err, "error copying connection.json")

	// Get metadata
	//metadata, err := getMetadata(metadataDir)
	//utils.HandleErr(err, "getting metadata for chaincode")
	tmpDir := "/tmp"
	execPath := path.Join(tmpDir, "exec")
	err = os.Mkdir(execPath, 777)
	utils.HandleErr(err, "error creating tmp exec directory")
	err = utils.Copy(chaincodeSourceDir, execPath)
	utils.HandleErr(err, "error copying to tmp exec directory")
	// copy metadata
	metadataDir := path.Join(chaincodeSourceDir, "metadata")
	metadataExists := utils.Exists(metadataDir)
	if metadataExists {
		metadataDestDir := path.Join(buildOutputDir, "metadata")
		err = utils.Copy(metadataDir, metadataDestDir)
		utils.HandleErr(err, "error copying metadata")
	}
	cfg, err := rest.InClusterConfig()
	utils.HandleErr(err, "error getting config from Kubernetes in the cluster")
	log.Printf("Config %v", cfg)
	clientSet, err := kubernetes.NewForConfig(cfg)
	utils.HandleErr(err, "error creating client set")
	ctx := context.Background()
	uid := shortuuid.New()[:6]
	podName := strings.ToLower(fmt.Sprintf("%s-%s", "nginx", uid))
	image := "dviejo/fabric-init:amd64-2.2.0"
	buildImage := "hyperledger/fabric-ccenv:2.2"
	fileServerIP := os.Getenv("FILE_SERVER_BASE_IP")
	fileServerURL := fmt.Sprintf("http://%s:8080", fileServerIP)
	mounts := []v1.VolumeMount{
		{
			Name:      "chaincode",
			MountPath: "/chaincode",
		},
	}
	pod, err := clientSet.CoreV1().Pods("default").Create(
		ctx,
		&v1.Pod{
			TypeMeta: v12.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: v12.ObjectMeta{
				Name: podName,
			},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: "chaincode",
					},
				},
				InitContainers: []v1.Container{
					// setup chaincode volume
					{
						Image:   image,
						Name:    "setup-chaincode-volume",
						Command: []string{"/bin/bash"},
						Args: []string{
							`-c`,
							`mkdir -p /chaincode/input /chaincode/output && chmod 777 /chaincode/input /chaincode/output`,
						},
						VolumeMounts: mounts,
					},
					// download chaincode source
					{
						Image:   image,
						Name:    "download-chaincode-source",
						Command: []string{"/bin/bash"},
						Args: []string{
							"-c",
							fmt.Sprintf(`curl -s -o- -L '%s/chaincode-source.tar' | tar -C /chaincode/input -xvf - && chmod -R 777 /chaincode/input`, fileServerURL),
						},
						VolumeMounts: mounts,
					},
					// build chaincode
					{
						Image:   buildImage,
						Name:    "build-go-chaincode",
						Command: []string{"/bin/bash"},
						Args: []string{
							"-c",
							`
set -e
if [ -x /chaincode/build.sh ]; then
/chaincode/build.sh
else
cp -R /chaincode/input/src/. /chaincode/output && cd /chaincode/output && npm install --production
fi
`,
						},
						VolumeMounts: mounts,
					},
				},
				// upload chaincode
				Containers: []v1.Container{
					{
						Name:         "upload-chaincode-output",
						Image:        image,
						VolumeMounts: mounts,
						Command:      []string{"/bin/bash"},
						Args: []string{
							"-c",
							fmt.Sprintf(
								`
cd /chaincode/output && tar cvf /chaincode/output.tar $(ls -A) && curl
-s --upload-file /chaincode/output.tar
'%s/build514067859/chaincode-output.tar`,
								fileServerURL,
							),
						},
					},
				},
			},
		},
		v12.CreateOptions{},
	)
	utils.HandleErr(err, "error creating pod")
	log.Printf("Pod created %s", pod.UID)

	defer cleanupPodSilent(clientSet, pod)
	// Watch builder Pod for completion or failure
	podSucceeded, err := watchPodUntilCompletion(ctx, clientSet, pod)
	utils.HandleErr(err, "watching builder pod")
	if !podSucceeded {
		fmt.Errorf("build of Chaincode in Pod %s failed", pod.Name)
	}
}

func cleanupPodSilent(clientset *kubernetes.Clientset, pod *v1.Pod) {
	err := cleanupPod(clientset, pod)
	log.Println(err)
}

func cleanupPod(clientset *kubernetes.Clientset, pod *v1.Pod) error {
	ctx := context.Background()
	return clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, v12.DeleteOptions{})
}

func watchPodUntilCompletion(ctx context.Context, clientset *kubernetes.Clientset, pod *v1.Pod) (bool, error) {
	/* Create log attacher
	var attachOnce sync.Once
	attachLogs := func() {
		go func() {
			err := streamPodLogs(pod)
			if err != nil {
				log.Printf("While streaming pod logs: %q", err)
			}
		}()
	}*/

	// Create informer
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace(pod.Namespace))
	informer := factory.Core().V1().Pods().Informer()
	c := make(chan struct{})
	defer close(c)

	podSuccessfull := make(chan bool)
	defer close(podSuccessfull)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldPod, newPod interface{}) {
			p := newPod.(*v1.Pod)
			if p.Name == pod.Name {
				log.Printf("Received update on pod %s, phase %s", p.Name, p.Status.Phase)
				// TODO: Can we miss an update, so not getting logs?

				switch p.Status.Phase {
				case v1.PodSucceeded:
					podSuccessfull <- true
				case v1.PodFailed, v1.PodUnknown:
					podSuccessfull <- false
				case v1.PodPending, v1.PodRunning:
					// Do nothing as this state is good
				default:
					podSuccessfull <- false // Unknown phase
				}
			}
		},
		DeleteFunc: func(oldPod interface{}) {
			p := oldPod.(*v1.Pod)
			if p.Name == pod.Name {
				log.Printf("Pod %s, phase %s got deleted", p.Name, p.Status.Phase)
				podSuccessfull <- false
			}
		},
	})
	go informer.Run(c)

	// Wait for result of informer and stop it afterwards.
	res := <-podSuccessfull
	c <- struct{}{}

	// Stream logs
	// TODO: This should be done as soon as the pod is running or has an result
	//err = streamPodLogs(ctx, pod)
	//if err != nil {
	//	log.Printf("While streaming pod logs: %q", err)
	//}

	return res, nil
}

// ChaincodeMetadata is based on
// https://github.com/hyperledger/fabric/blob/v2.0.1/core/chaincode/persistence/chaincode_package.go#L226
type ChaincodeMetadata struct {
	Type       string `json:"type"` // golang, java, node
	Path       string `json:"path"`
	Label      string `json:"label"`
	MetadataID string
}

func getMetadata(metadataDir string) (*ChaincodeMetadata, error) {
	metadataFile := filepath.Join(metadataDir, "metadata.json")
	metadataData, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		return nil, errors.Wrap(err, "Reading metadata.json")
	}

	metadata := ChaincodeMetadata{}
	err = json.Unmarshal(metadataData, &metadata)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshaling metadata.json")
	}

	// Create hash in order to track this CC
	h := sha1.New() // #nosec G401
	_, err = h.Write(metadataData)
	if err != nil {
		return nil, errors.Wrap(err, "hashing metadata")
	}

	metadata.MetadataID = fmt.Sprintf("%x", h.Sum(nil))[0:8]

	return &metadata, nil
}
