package empty

import (
	_ "fmt"
	_ "log"
	_ "net/http"

	_ "github.com/dchest/uniuri"
	_ "github.com/gorilla/mux"
	_ "github.com/minio/minio-go/v6"
	_ "k8s.io/api/batch/v1"
	_ "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/rest"
)
