module github.com/alkaid/behavior

go 1.18

require (
	github.com/alkaid/timingwheel v1.0.2
	github.com/bilibili/gengine v1.5.7
	github.com/disiqueira/gotree v1.0.0
	github.com/google/go-cmp v0.5.8
	github.com/matoous/go-nanoid v1.5.0
	github.com/panjf2000/ants/v2 v2.5.0
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.25.0
	go.uber.org/zap v1.21.0
)

require (
	github.com/antlr/antlr4 v0.0.0-20210105192202-5c2b686f95e1 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/exp v0.0.0-20220706164943-b4a6d9510983 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/panjf2000/ants/v2 v2.5.0 => github.com/alkaid/ants/v2 v2.5.101

//replace github.com/panjf2000/ants/v2 => ../ants
