package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	create          = kingpin.Command("create", "Create new Deployment.")
	user            = create.Flag("user", "User Id for Kubernetes").Default("admin").String()
	pw              = create.Flag("password", "Password for Kubernetes").Required().String()
	url             = create.Flag("url", "API Url").Required().String()
	namespace       = create.Flag("ns", "Namespace").Default("default").String()
	imagepullpolicy = create.Flag("imagepullpolicy", "Image Pull Policy").Default("Always").String()
	deploymentName  = create.Flag("name", "Deployment Name").Required().String()
	replicas        = create.Flag("replicas", "No. of replicas").Default("1").Int()
	image           = create.Flag("image", "Docker Image").Required().String()
	port            = create.Flag("port", "container ports").Required().Int()
	imagepullsecret = create.Flag("imagepullsecret", "Image Pull Secret").Required().String()
	secret          = create.Flag("secret", "Set secret name and mountpath").StringMap()
	env             = create.Flag("env", "Add env variable and it's value").StringMap()
	limit           = create.Flag("limit", "Add memory limit").String()
	request         = create.Flag("request", "Add memory request").String()

	replace          = kingpin.Command("replace", "Replace a current Deployment.")
	userR            = replace.Flag("user", "User Id for Kubernetes").Default("admin").String()
	pwR              = replace.Flag("password", "Password for Kubernetes").Required().String()
	urlR             = replace.Flag("url", "API Url").Required().String()
	namespaceR       = replace.Flag("ns", "Namespace").Default("default").String()
	imagepullpolicyR = replace.Flag("imagepullpolicy", "Image Pull Policy").Default("Always").String()
	deploymentNameR  = replace.Flag("name", "Deployment Name").Required().String()
	replicasR        = replace.Flag("replicas", "No. of replicas").Default("1").Int()
	imageR           = replace.Flag("image", "Docker Image").Required().String()
	portR            = replace.Flag("port", "container ports").Required().Int()
	imagepullsecretR = replace.Flag("imagepullsecret", "Image Pull Secret").Required().String()
	secretR          = replace.Flag("secret", "Set secret name and mountpath").StringMap()
	envR             = replace.Flag("env", "Add env variable and it's value").StringMap()
	limitR           = replace.Flag("limit", "Add memory limit").String()
	requestR         = replace.Flag("request", "Add memory request").String()

	patch            = kingpin.Command("patch", "Patch a current Deployment.")
	userP            = patch.Flag("user", "User Id for Kubernetes").Default("admin").String()
	pwP              = patch.Flag("password", "Password for Kubernetes").Required().String()
	urlP             = patch.Flag("url", "API Url").Required().String()
	namespaceP       = patch.Flag("ns", "Namespace").Default("default").String()
	imagepullpolicyP = patch.Flag("imagepullpolicy", "Image Pull Policy").Default("Always").String()
	deploymentNameP  = patch.Flag("name", "Deployment Name").Required().String()
	replicasP        = patch.Flag("replicas", "No. of replicas").Default("1").Int()
	imageP           = patch.Flag("image", "Docker Image").Required().String()
	portP            = patch.Flag("port", "container ports").Required().Int()
	imagepullsecretP = patch.Flag("imagepullsecret", "Image Pull Secret").Required().String()
	secretP          = patch.Flag("secret", "Set secret name and mountpath").StringMap()
	envP             = patch.Flag("env", "Add env variable and it's value").StringMap()
	limitP           = patch.Flag("limit", "Add memory limit").String()
	requestP         = patch.Flag("request", "Add memory request").String()

	del             = kingpin.Command("delete", "Delete a current Deployment.")
	userD           = del.Flag("user", "User Id for Kubernetes").Default("admin").String()
	pwD             = del.Flag("password", "Password for Kubernetes").Required().String()
	urlD            = del.Flag("url", "API Url").Required().String()
	namespaceD      = del.Flag("ns", "Namespace").Default("default").String()
	deploymentNameD = del.Flag("name", "Deployment Name").Required().String()
)

type Deployment struct {
	ApiVersion string         `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Metadata   *Meta          `json:"metadata,omitempty"`
	Spec       DeploymentSpec `json:"spec"`
}
type Meta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
type DeploymentSpec struct {
	Replicas int           `json:"replicas"`
	Template PodTemplate   `json:"template"`
	Selector LabelSelector `json:"selector"`
}
type LabelSelector struct {
	MatchLabels struct {
		App string `json:"app"`
	} `json:"matchLabels"`
}
type PodTemplate struct {
	Metadata *ObjMeta `json:"metadata,omitempty"`
	Spec     PodSpec  `json:"spec"`
}
type ObjMeta struct {
	Labels struct {
		App string `json:"app"`
	} `json:"labels"`
}
type Secret struct {
	SecretName string `json:"secretName"`
}
type Volume struct {
	Name   string `json:"name"`
	Secret Secret `json:"secret,omitempty"`
	Pvc    *Pvc   `json:"persistentVolumeClaim,omitempty"`
}
type Pvc struct {
	ClaimName string `json:"claimName,omitempty"`
}
type PodSpec struct {
	Containers       []Container `json:"containers"`
	Volumes          []Volume    `json:"volumes"`
	ImagePullSecrets []struct {
		Name string `json:"name"`
	} `json:"imagePullSecrets"`
}
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"readOnly"`
}

type Container struct {
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy"`
	Name            string `json:"name"`
	Ports           []struct {
		ContainerPort int `json:"containerPort"`
	} `json:"ports"`
	VolumeMounts []VolumeMount `json:"volumeMounts"`
	Env          []EnvVar      `json:"env"`
	Resources    Resources     `json:"resources"`
}
type Resources struct {
	Limits   *Memory `json:"limits"`
	Requests *Memory `json:"requests"`
}
type Memory struct {
	Memory string `json:"memory"`
}

func main() {
	code := 0
	switch kingpin.Parse() {
	case "create":
		code = createDeployment()
	case "delete":
		code = deleteDeployment()
	case "replace":
		code = replaceDeployment()
	case "patch":
		code = patchDeployment()
	}
	os.Exit(code)
}
func createDeployment() int {
	d := Deployment{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: &Meta{
			Name:      *deploymentName,
			Namespace: *namespace,
		},
		Spec: DeploymentSpec{
			Replicas: *replicas,
			Selector: LabelSelector{
				MatchLabels: struct {
					App string `json:"app"`
				}{
					App: *deploymentName,
				},
			},
			Template: PodTemplate{
				Metadata: &ObjMeta{
					struct {
						App string `json:"app"`
					}{
						App: *deploymentName,
					},
				},
				Spec: PodSpec{
					Containers: []Container{
						{
							Image:           *image,
							ImagePullPolicy: *imagepullpolicy,
							Name:            *deploymentName,
							Ports: []struct {
								ContainerPort int `json:"containerPort"`
							}{
								{
									ContainerPort: *port,
								},
							},
						},
					},
					ImagePullSecrets: []struct {
						Name string `json:"name"`
					}{
						{
							Name: *imagepullsecret,
						},
					},
				},
			},
		},
	}

	d.Spec.Template.Spec.Containers[0].VolumeMounts = make([]VolumeMount, len(*secret))
	d.Spec.Template.Spec.Volumes = make([]Volume, len(*secret))
	j := 0
	for k, v := range *secret {
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].Name = k
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].MountPath = v
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].ReadOnly = true
		d.Spec.Template.Spec.Volumes[j].Name = k
		d.Spec.Template.Spec.Volumes[j].Secret.SecretName = k
		j++
	}

	d.Spec.Template.Spec.Containers[0].Env = make([]EnvVar, len(*env))
	i := 0
	for k, v := range *env {
		d.Spec.Template.Spec.Containers[0].Env[i].Name = k
		d.Spec.Template.Spec.Containers[0].Env[i].Value = v
		i++
	}

	if len(*limit) > 0 {
		d.Spec.Template.Spec.Containers[0].Resources.Limits = &Memory{*limit}
	}
	if len(*request) > 0 {
		d.Spec.Template.Spec.Containers[0].Resources.Requests = &Memory{*request}
	}

	out, err := json.Marshal(d)
	if err != nil {
		fmt.Printf("err : %v", err)
	}

	_, err = send(*url+"/apis/apps/v1/namespaces/"+*namespace+"/deployments", *user, *pw, out, "POST")
	if err != nil {
		return -1
	}
	return 0
}
func deleteDeployment() int {
	send(*urlD+"/apis/apps/v1/namespaces/"+*namespaceD+"/deployments/"+*deploymentNameD, *userD, *pwD, []byte(""), "DELETE")
	list, err := send(*urlD+"/apis/apps/v1/namespaces/"+*namespaceD+"/replicasets", *userD, *pwD, []byte(""), "GET")
	if err != nil {
		return -2
	}
	var name string
	found := false
	if list == nil || list["items"] == nil {
		return -2
	}
	replicasets := list["items"].([]interface{})
	for _, repli := range replicasets {
		meta := repli.(map[string]interface{})["metadata"].(map[string]interface{})
		labels := meta["labels"].(map[string]interface{})
		app := labels["app"].(string)
		if app == *deploymentNameD {
			name = meta["name"].(string)
			found = true
			break
		}
	}
	if !found {
		fmt.Println("Replicaset for " + *deploymentNameD + " not found")
		return -2
	}
	_, err = send(*urlD+"/apis/apps/v1/namespaces/"+*namespaceD+"/replicasets/"+name, *userD, *pwD, []byte(""), "DELETE")
	if err != nil {
		return -1
	}
	return 0
	// send(*urlD+"/api/v1/namespaces/"+*namespaceD+"/pods", *userD, *pwD, []byte(""), "DELETE")
}
func replaceDeployment() int {
	d := Deployment{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: &Meta{
			Name:      *deploymentNameR,
			Namespace: *namespaceR,
		},
		Spec: DeploymentSpec{
			Replicas: *replicasR,
			Selector: LabelSelector{
				MatchLabels: struct {
					App string `json:"app"`
				}{
					App: *deploymentName,
				},
			},
			Template: PodTemplate{
				Metadata: &ObjMeta{
					struct {
						App string `json:"app"`
					}{
						App: *deploymentNameR,
					},
				},
				Spec: PodSpec{
					Containers: []Container{
						{
							Image:           *imageR,
							ImagePullPolicy: *imagepullpolicyR,
							Name:            *deploymentNameR,
							Ports: []struct {
								ContainerPort int `json:"containerPort"`
							}{
								{
									ContainerPort: *portR,
								},
							},
						},
					},
					ImagePullSecrets: []struct {
						Name string `json:"name"`
					}{
						{
							Name: *imagepullsecretR,
						},
					},
				},
			},
		},
	}
	d.Spec.Template.Spec.Containers[0].VolumeMounts = make([]VolumeMount, len(*secretR))
	d.Spec.Template.Spec.Volumes = make([]Volume, len(*secretR))
	j := 0
	for k, v := range *secretR {
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].Name = k
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].MountPath = v
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].ReadOnly = true
		d.Spec.Template.Spec.Volumes[j].Name = k
		d.Spec.Template.Spec.Volumes[j].Secret.SecretName = k
		j++
	}

	d.Spec.Template.Spec.Containers[0].Env = make([]EnvVar, len(*envR))
	i := 0
	for k, v := range *envR {
		d.Spec.Template.Spec.Containers[0].Env[i].Name = k
		d.Spec.Template.Spec.Containers[0].Env[i].Value = v
		i++
	}

	if len(*limitR) > 0 {
		d.Spec.Template.Spec.Containers[0].Resources.Limits = &Memory{*limitR}
	}
	if len(*requestR) > 0 {
		d.Spec.Template.Spec.Containers[0].Resources.Requests = &Memory{*requestR}
	}

	out, err := json.Marshal(d)
	if err != nil {
		fmt.Printf("err : %v", err)
	}
	_, err = send(*urlR+"/apis/apps/v1/namespaces/"+*namespaceR+"/deployments/"+*deploymentNameR, *userR, *pwR, out, "PUT")
	if err != nil {
		return -1
	}
	return 0
}

func patchDeployment() int {
	d := Deployment{
		Spec: DeploymentSpec{
			Replicas: *replicasP,
			Selector: LabelSelector{
				MatchLabels: struct {
					App string `json:"app"`
				}{
					App: *deploymentNameP,
				},
			},
			Template: PodTemplate{
				Metadata: &ObjMeta{
					struct {
						App string `json:"app"`
					}{
						App: *deploymentNameP,
					},
				},
				Spec: PodSpec{
					Containers: []Container{
						{
							Image:           *imageP,
							ImagePullPolicy: *imagepullpolicyP,
							Name:            *deploymentNameP,
							Ports: []struct {
								ContainerPort int `json:"containerPort"`
							}{
								{
									ContainerPort: *portP,
								},
							},
						},
					},
					ImagePullSecrets: []struct {
						Name string `json:"name"`
					}{
						{
							Name: *imagepullsecretP,
						},
					},
				},
			},
		},
	}

	d.Spec.Template.Spec.Containers[0].VolumeMounts = make([]VolumeMount, len(*secretP))
	d.Spec.Template.Spec.Volumes = make([]Volume, len(*secretP))
	j := 0
	for k, v := range *secretP {
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].Name = k
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].MountPath = v
		d.Spec.Template.Spec.Containers[0].VolumeMounts[j].ReadOnly = true
		d.Spec.Template.Spec.Volumes[j].Name = k
		d.Spec.Template.Spec.Volumes[j].Secret.SecretName = k
		j++
	}

	d.Spec.Template.Spec.Containers[0].Env = make([]EnvVar, len(*envP))
	i := 0
	for k, v := range *envP {
		d.Spec.Template.Spec.Containers[0].Env[i].Name = k
		d.Spec.Template.Spec.Containers[0].Env[i].Value = v
		i++
	}
	if len(*limitP) > 0 {
		d.Spec.Template.Spec.Containers[0].Resources.Limits = &Memory{*limitP}
	}
	if len(*requestP) > 0 {
		d.Spec.Template.Spec.Containers[0].Resources.Requests = &Memory{*requestP}
	}
	out, err := json.Marshal(d)
	if err != nil {
		fmt.Printf("err : %v", err)
	}
	// fmt.Println(string(out))
	_, err = send(*urlP+"/apis/apps/v1/namespaces/"+*namespaceP+"/deployments/"+*deploymentNameP, *userP, *pwP, out, "PATCH")
	if err != nil {
		return -1
	}
	return 0
}

func send(url string, user string, pw string, jsonStr []byte, method string) (map[string]interface{}, error) {
	// fmt.Println("URL:>", url)
	// println(string(jsonStr))
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
	req.SetBasicAuth(user, pw)
	if method != "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/merge-patch+json")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(buf, &result)

	fmt.Println("response Status:", resp.Status)
	if resp.StatusCode >= 300 {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("error reading response body")
		}
		bodyString := string(bodyBytes)
		println(bodyString)

		return nil, fmt.Errorf("call failed")
	}
	return result, nil
}
