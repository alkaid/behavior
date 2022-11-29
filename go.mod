module github.com/alkaid/behavior

go 1.19

require (
	github.com/alkaid/timingwheel v1.0.3
	github.com/disiqueira/gotree v1.0.0
	github.com/google/go-cmp v0.5.9
	github.com/matoous/go-nanoid v1.5.0
	github.com/panjf2000/ants/v2 v2.6.0
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.36.0
	go.uber.org/zap v1.23.0
)

require (
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/exp v0.0.0-20221126150942-6ab00d035af9 // indirect
)

replace github.com/panjf2000/ants/v2 v2.6.0 => github.com/alkaid/ants/v2 v2.6.100

//replace github.com/panjf2000/ants/v2 => ../ants
