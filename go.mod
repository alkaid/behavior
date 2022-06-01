module github.com/alkaid/behavior

go 1.18

require (
	github.com/alkaid/timingwheel v1.0.2
	github.com/panjf2000/ants/v2 v2.4.8
	github.com/samber/lo v1.21.0
	go.uber.org/zap v1.21.0
)

require (
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/exp v0.0.0-20220303212507-bbda1eaf7a17 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/panjf2000/ants/v2 v2.4.8 => github.com/alkaid/ants/v2 v2.4.802
