package tracer

import (
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/iotexproject/iotex-core/v2/pkg/version"
)

const (
	_service = "iotex-tracer"
)

// Config is the config for tracer
type Config struct {
	// ServiceName customize service name
	ServiceName string `yaml:"serviceName"`
	// EndPoint the jaeger endpoint
	EndPoint string `yaml:"endpoint"`
	// InstanceID MUST be unique for each instance of the same
	InstanceID string `yaml:"instanceID"`
	//SamplingRatio customize sampling ratio
	//ratio >= 1 will always sample (default),< 0 are treated as zero will no sample
	// if you set this to .5, half of traces will be sampled
	SamplingRatio string `yaml:"samplingRatio"`
}

// Option the tracer provider option
type Option func(ops *optionParams) error

type optionParams struct {
	serviceName   string
	endpoint      string //the jaeger endpoint
	instanceID    string //Note: MUST be unique for each instance of the same
	samplingRatio string
}

// WithServiceName defines service name
func WithServiceName(name string) Option {
	return func(ops *optionParams) error {
		ops.serviceName = name
		return nil
	}
}

// WithEndpoint defines the full URL to the Jaeger HTTP Thrift collector
func WithEndpoint(endpoint string) Option {
	return func(ops *optionParams) error {
		ops.endpoint = endpoint
		return nil
	}
}

// WithInstanceID defines the instance id
func WithInstanceID(instanceID string) Option {
	return func(ops *optionParams) error {
		ops.instanceID = instanceID
		return nil
	}
}

// WithSamplingRatio defines the sampling ratio
func WithSamplingRatio(samplingRatio string) Option {
	return func(ops *optionParams) error {
		ops.samplingRatio = samplingRatio
		return nil
	}
}

// NewProvider create an instance of tracer provider
func NewProvider(opts ...Option) (*tracesdk.TracerProvider, error) {
	var (
		err                           error
		ops                           optionParams
		trackerTracerProviderOption   []tracesdk.TracerProviderOption
		jaegerCollectorEndpointOption []jaeger.CollectorEndpointOption
	)
	for _, opt := range opts {
		if err = opt(&ops); err != nil {
			return nil, err
		}
	}
	if ops.endpoint != "" {
		jaegerCollectorEndpointOption = append(jaegerCollectorEndpointOption, jaeger.WithEndpoint(ops.endpoint))
	} else {
		//skipped tracing when endpoint no set
		return nil, nil
	}
	if ops.serviceName == "" {
		ops.serviceName = _service
	}
	if ops.samplingRatio != "" {
		ratio, err := strconv.ParseFloat(ops.samplingRatio, 64)
		if err != nil {
			return nil, err
		}
		trackerTracerProviderOption = append(trackerTracerProviderOption,
			tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(ratio))),
		)
	}
	kv := []attribute.KeyValue{
		semconv.ServiceVersionKey.String(version.PackageVersion),
		semconv.ServiceNameKey.String(ops.serviceName),
	}
	if ops.instanceID != "" {
		kv = append(kv, semconv.ServiceInstanceIDKey.String(ops.instanceID))
	}
	// Record information about this application in an Resource.
	resources := tracesdk.WithResource(resource.NewWithAttributes(
		semconv.SchemaURL,
		kv...,
	))
	trackerTracerProviderOption = append(trackerTracerProviderOption, resources)
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaegerCollectorEndpointOption...))
	if err != nil {
		return nil, err
	}
	// Always be sure to batch in production.
	trackerTracerProviderOption = append(trackerTracerProviderOption, tracesdk.WithBatcher(exp))
	tp := tracesdk.NewTracerProvider(trackerTracerProviderOption...)
	//set global provider
	otel.SetTracerProvider(tp)
	return tp, nil
}
