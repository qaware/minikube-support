module github.com/chr-fritz/minikube-support

go 1.12

require (
	github.com/buger/goterm v0.0.0-20181115115552-c206103e1f37
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.1
	github.com/hashicorp/go-multierror v1.0.0
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/miekg/dns v1.1.6
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/net v0.0.0-20190514140710-3ec191127204 // indirect
	google.golang.org/grpc v1.20.1
)

replace github.com/golang/glog => github.com/kubermatic/glog-logrus v0.0.0-20180829085450-3fa5b9870d1d
