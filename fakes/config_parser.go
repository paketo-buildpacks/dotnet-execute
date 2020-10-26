package fakes

import (
	"sync"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
)

type ConfigParser struct {
	ParseCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			RuntimeConfig dotnetexecute.RuntimeConfig
			Error         error
		}
		Stub func(string) (dotnetexecute.RuntimeConfig, error)
	}
}

func (f *ConfigParser) Parse(param1 string) (dotnetexecute.RuntimeConfig, error) {
	f.ParseCall.Lock()
	defer f.ParseCall.Unlock()
	f.ParseCall.CallCount++
	f.ParseCall.Receives.Path = param1
	if f.ParseCall.Stub != nil {
		return f.ParseCall.Stub(param1)
	}
	return f.ParseCall.Returns.RuntimeConfig, f.ParseCall.Returns.Error
}
