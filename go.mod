module github.com/qaware/minikube-support

go 1.14

require (
	github.com/buger/goterm v0.0.0-20200322175922-2f3e71b85129
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/magiconair/properties v1.8.1
	github.com/miekg/dns v1.1.29
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	google.golang.org/grpc v1.30.0
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.0
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19 // indirect
)

replace github.com/golang/glog => github.com/kubermatic/glog-logrus v0.0.0-20180829085450-3fa5b9870d1d
