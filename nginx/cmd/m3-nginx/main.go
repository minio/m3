package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func getK8sConfig() *rest.Config {
	// creates the in-cluster config
	var config *rest.Config
	if os.Getenv("DEVELOPMENT") != "" {
		//when doing local development, mount k8s api via `kubectl proxy`
		config = &rest.Config{
			Host:            "http://localhost:8001",
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			APIPath:         "/",
			BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6InFETTJ6R21jMS1NRVpTOER0SnUwdVg1Q05XeDZLV2NKVTdMUnlsZWtUa28ifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRldi1zYS10b2tlbi14eGxuaiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJkZXYtc2EiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJmZDVhMzRjNy0wZTkwLTQxNTctYmY0Zi02Yjg4MzIwYWIzMDgiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkZXYtc2EifQ.woZ6Bmkkw-BMV-_UX0Y-S_Lkb6H9zqKZX2aNhyy7valbYIZfIzrDqJYWV9q2SwCP20jBfdsDS40nDcMnHJPE5jZHkTajAV6eAnoq4EspRqORtLGFnVV-JR-okxtvhhQpsw5MdZacJk36ED6Hg8If5uTOF7VF5r70dP7WYBMFiZ3HSlJBnbu7QoTKFmbJ1MafsTQ2RBA37IJPkqi3OHvPadTux6UdMI8LlY7bLkZkaryYR36kwIzSqsYgsnefmm4eZkZzpCeyS9scm9lPjeyQTyCAhftlxfw8m_fsV0EDhmybZCjgJi4R49leJYkHdpnCSkubj87kJAbGMwvLhMhFFQ",
		}
	} else {
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

	}

	return config
}

// main retrieves the initial nginx configuration from configmap and then starts nginx, if this is successful it then
// starts the main loop of monitoring for configMap changes
func main() {
	log.Println("Starting ConfigMap Watcher.")
	clientSet, err := kubernetes.NewForConfig(getK8sConfig())
	if err != nil {
		panic(err)
	}

	// Load config and start nginx

	cfgMap, err := clientSet.CoreV1().ConfigMaps("default").Get("nginx-configuration", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	if val, ok := cfgMap.Data["nginx.conf"]; ok {
		go startNginx(val)
	}

	log.Println("Done Starting nginx")

	// informer factory
	doneCh := make(chan struct{})
	factory := informers.NewSharedInformerFactory(clientSet, 0)

	log.Println("Start informer")

	cfgInformer := factory.Core().V1().ConfigMaps().Informer()
	cfgInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			cfgMap := newObj.(*v1.ConfigMap)
			log.Println(cfgMap.Name)
			// monitor for nginx config map
			if strings.HasPrefix(cfgMap.ObjectMeta.Name, "nginx") {
				log.Println("nginx configMap updated:", cfgMap.Name)
				if val, ok := cfgMap.Data["nginx.conf"]; ok {
					go reloadNginx(val)
				}
			}
		},
	})

	go cfgInformer.Run(doneCh)
	//block until the informer exits
	<-doneCh

	log.Println("informer complete")

}

// writeNginxConf writes the new config to the nginx.conf file for nginx to consume
func writeNginxConf(config string) {
	err := ioutil.WriteFile("etc/nginx/nginx.conf", []byte(config), 0644)
	if err != nil {
		log.Println(err)
	}
}

// startNginx starts the nginx process and redirects all it's output to os.Stdout
func startNginx(config string) {
	writeNginxConf(config)
	log.Println("Starting Nginx")
	cmd := exec.Command("nginx", "-g", "daemon off;")
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}

func reloadNginx(config string) {
	writeNginxConf(config)
	log.Println("Reloading Nginx")
	cmd := exec.Command("nginx", "-s", "reload")
	err := cmd.Run()
	if err != nil {
		log.Println("cmd.Run() failed with %s\n", err)
	}
}
