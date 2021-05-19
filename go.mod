module github.com/peytoncasper/waypoint-plugin-spark

go 1.14

require (
	cloud.google.com/go/storage v1.10.0
	github.com/hashicorp/waypoint-plugin-sdk v0.0.0-20210510195008-b42c688ebedf
	google.golang.org/api v0.46.0 // indirect
	google.golang.org/genproto v0.0.0-20210510173355-fb37daa5cd7a // indirect
	google.golang.org/protobuf v1.26.0
	cloud.google.com/go v0.81.0
)

// replace github.com/hashicorp/waypoint-plugin-sdk => ../../waypoint-plugin-sdk
