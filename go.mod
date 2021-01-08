module github.com/qaware/minikube-support

go 1.14

require (
	github.com/awesome-gocui/gocui v0.6.1-0.20200808231733-d0eae9ef0497
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.3
	github.com/golangci/golangci-lint v1.35.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.1
	github.com/miekg/dns v1.1.29
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/sys v0.0.0-20210105210732-16f7687f5001 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20210106152847-07624b53cd92 // indirect
	google.golang.org/grpc v1.34.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
)

replace github.com/golang/glog => github.com/kubermatic/glog-logrus v0.0.0-20180829085450-3fa5b9870d1d

replace github.com/awesome-gocui/termbox-go => github.com/nsf/termbox-go v0.0.0-20201124104050-ed494de23a00
