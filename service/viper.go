package service

import (
	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/zk"
	"github.com/spf13/viper"
)

const (
	// DiscoveryKey is the default Viper subkey used for service discovery configuration.
	// WebPA servers should typically use this key as a standard.
	DiscoveryKey = "discovery"
)

// NewOptions produces an Options from a Viper instance.  Typically, the Viper instance
// will be configured via the server package.
//
// Since service discovery is an optional module for a WebPA server, this function allows
// the supplied Viper to be nil or otherwise uninitialized.  Client code that opts in to
// service discovery can thus use the same codepath to configure an Options instance.
func NewOptions(logger logging.Logger, pingFunc func() error, v *viper.Viper) (o *Options, err error) {
	o = new(Options)
	if v != nil {
		err = v.Unmarshal(o)
	}

	o.Logger = logger
	o.PingFunc = pingFunc
	return
}

// Initialize is the top-level function for bootstrapping the service discovery infrastructure
// using a Viper instance.  No watches are set by this function, but all registrations are made
// and monitored via the returned RegistrarWatcher.
func Initialize(logger logging.Logger, pingFunc func() error, v *viper.Viper) (o *Options, r Registrar, re RegisteredEndpoints, err error) {
	o, err = NewOptions(logger, pingFunc, v)
	if err != nil {
		return
	}

	r = NewRegistrar(o)
	re, err = RegisterAll(r, o)
	return
}

func NewGokitRegistrars(o *Options, logger log.Logger) ([]sd.Registrar, error) {
	registrations := o.registrations()
	if len(registrations) == 0 {
		return nil, nil
	}

	// TODO: Just log to stdout for now until we fully migrate to go-kit
	client, err := zk.NewClient(
		o.servers(),
		logger,
		zk.ConnectTimeout(o.connectTimeout()),
	)

	if err != nil {
		return nil, err
	}

	registrars := make([]sd.Registrar, 0, len(registrations))
	for _, registration := range registrations {
		r := zk.NewRegistrar(
			client,
			zk.Service{
				Path: o.baseDirectory(),
				Name: o.serviceName(),
				Data: []byte(registration),
			},
			logger,
		)

		r.Register()
		registrars = append(registrars, r)
	}

	return registrars, nil
}

func NewGokitInstancers(o *Options, logger log.Logger) ([]sd.Instancer, error) {
	return nil, nil
}
