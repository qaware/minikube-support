module github.com/chr-fritz/minikube-support

go 1.12

require (
	github.com/hashicorp/go-multierror v1.0.0
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
)

replace github.com/golang/glog => github.com/kubermatic/glog-logrus v0.0.0-20180829085450-3fa5b9870d1d
