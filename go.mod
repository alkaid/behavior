module github.com/alkaid/behavior

go 1.24.2

require (
	github.com/alkaid/timingwheel v1.0.5
	github.com/disiqueira/gotree v1.0.0
	github.com/google/go-cmp v0.7.0
	github.com/matoous/go-nanoid v1.5.1
	github.com/panjf2000/ants/v2 v2.11.3
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.50.0
	go.uber.org/zap v1.27.0
)

require (
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)

replace github.com/panjf2000/ants/v2 v2.11.3 => github.com/alkaid/ants/v2 v2.11.3-ak-2

//replace github.com/panjf2000/ants/v2 => ../ants
