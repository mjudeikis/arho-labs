package workers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mjudeikis/osa-labs/pkg/api"
	"github.com/mjudeikis/osa-labs/pkg/store"
	"github.com/mjudeikis/osa-labs/pkg/utils/keygen"
	"github.com/mjudeikis/osa-labs/pkg/utils/random"
	"github.com/mjudeikis/osa-labs/pkg/utils/wait"
)

const workersNamespace = "workers"

type kubeWorkers struct {
	client kubernetes.Interface
	sync.Mutex
	log    *logrus.Entry
	image  string
	number int
	store  store.Store

	dCli      appsv1client.DeploymentInterface
	svcCli    corev1client.ServiceInterface
	secretCli corev1client.SecretInterface
}

var _ Workers = &kubeWorkers{}

func (k kubeWorkers) Get() (*api.Worker, error) {

	return nil, nil
}

func (k kubeWorkers) Create() error {
	deploymentList, err := k.dCli.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	n := k.number - len(deploymentList.Items)
	k.log.Infof("create workers %v", n)
	if n > 0 {
		for i := 0; i < n; i++ {
			name, err := k.createWorker()
			if err != nil {
				return err
			}
			k.log.Println(name)
		}
	}

	return k.reconcileWorkers(context.Background(), k.number )
}

func New(log *logrus.Entry, image string, number int) (Workers, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	storage, err := store.New(log, "storage", "credentials")
	if err != nil {
		return nil, err
	}

	return &kubeWorkers{
		log:       log,
		client:    cli,
		image:     image,
		number:    number,
		store:     storage,
		dCli:      cli.AppsV1().Deployments(workersNamespace),
		svcCli:    cli.CoreV1().Services(workersNamespace),
		secretCli: cli.CoreV1().Secrets(workersNamespace),
	}, nil
}

func (k kubeWorkers) reconcileWorkers(ctx context.Context, num int) error {

	err := wait.PollImmediateUntil(10*time.Second, func() (bool, error) {
		ready := 0
		deploymentList, err := k.dCli.List(metav1.ListOptions{})
		if err != nil {
			k.log.Error(err)
			return false, nil
		}
		for _, dc := range deploymentList.Items {
			if dc.Status.Replicas == dc.Status.ReadyReplicas {
				ready++
				k.log.Infof("deployment count: %v . Ready %v. Expected: %v", len(deploymentList.Items), ready, num)
			}
		}
		if ready >= num {
			k.log.Info("all deployment ready")
			return true, nil
		}
		k.log.Debug("reconcile workers")
		return false, nil
	}, ctx.Done())

	err = wait.PollImmediateUntil(10*time.Second, func() (bool, error) {
		ready := 0
		svcList, err := k.svcCli.List(metav1.ListOptions{})
		if err != nil {
			k.log.Error(err)
			return false, nil
		}
		for _, svc := range svcList.Items {
			if len(svc.Status.LoadBalancer.Ingress) >= 1 {
				ready++
				k.log.Infof("svc count: %v . Ready %v. Expected: %v", len(svcList.Items), ready, num)
			}
		}
		if ready >= num {
			k.log.Info("all services ready")
			return true, nil
		}
		k.log.Debug("reconcile IP's")
		return false, nil
	}, ctx.Done())

	// populate database
	var workerStore api.WorkersStore
	deploymentList, err := k.dCli.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, dc := range deploymentList.Items {
		k.log.Debug(dc.GetName())
		secret, err := k.secretCli.Get(dc.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		sshKey := string(secret.Data["id_rsa"][:])

		svc, err := k.svcCli.Get(dc.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		ip := svc.Status.LoadBalancer.Ingress[0].IP

		workerStore.Workers = append(workerStore.Workers, api.Worker{
			Name:     dc.GetName(),
			IP:       ip,
			Reserved: false,
			SSHKey:   sshKey,
		})
	}

	bytes, err := yaml.Marshal(workerStore)
	if err != nil {
		return err
	}

	return k.store.Put("workers", bytes)
}

func (k kubeWorkers) createWorker() (string, error) {
	template, err := getWorkerTemplate(k.image)
	if err != nil {
		return "", err
	}

	_, err = k.dCli.Create(template["deployment"].(*appsv1.Deployment))
	if err != nil {
		return "", err
	}
	_, err = k.svcCli.Create(template["service"].(*apiv1.Service))
	if err != nil {
		return "", err
	}
	_, err = k.secretCli.Create(template["secret"].(*apiv1.Secret))
	if err != nil {
		return "", err
	}

	return template["deployment"].(*appsv1.Deployment).GetName(), nil
}

func getWorkerTemplate(image string) (map[string]interface{}, error) {
	name, err := random.LowerCaseAlphaString(10)
	if err != nil {
		return nil, err
	}

	template := make(map[string]interface{})

	dt := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"worker": "kube",
					"name":   name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"worker": "kube",
						"name":   name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "worker",
							Image:           image,
							ImagePullPolicy: apiv1.PullAlways,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "ssh",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 2222,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "sshkey",
									ReadOnly:  true,
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "sshkey",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: name,
								},
							},
						},
					},
				},
			},
		},
	}

	svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "ssh",
					Port: 2222,
				},
			},
			Type: apiv1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"name": name,
			},
		},
	}

	sshKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privateKeyByte, err := keygen.PrivateKeyAsBytes(sshKey)
	if err != nil {
		return nil, err
	}
	publicKeyString, err := keygen.SSHPublicKeyAsString(&sshKey.PublicKey)
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{
		"id_rsa":     privateKeyByte,
		"id_rsa.pub": []byte(publicKeyString),
	}
	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: data,
	}

	template["deployment"] = dt
	template["service"] = svc
	template["secret"] = secret

	return template, nil
}

func int32Ptr(i int32) *int32 { return &i }

func getConfig() (*rest.Config, error) {
	if os.Getenv("KUBECONFIG") != "" {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	} else {
		return rest.InClusterConfig()
	}
}
