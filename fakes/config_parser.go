package fakes

import (
	"sync"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
)

type ConfigParser struct {
	ParseCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Glob string
		}
		Returns struct {
			RuntimeConfig dotnetexecute.RuntimeConfig
			Error         error
		}
		Stub func(string) (dotnetexecute.RuntimeConfig, error)
	}
}

func (f *ConfigParser) Parse(param1 string) (dotnetexecute.RuntimeConfig, error) {
	f.ParseCall.mutex.Lock()
	defer f.ParseCall.mutex.Unlock()
	f.ParseCall.CallCount++
	f.ParseCall.Receives.Glob = param1
	if f.ParseCall.Stub != nil {
		return f.ParseCall.Stub(param1)
	}
	return f.ParseCall.Returns.RuntimeConfig, f.ParseCall.Returns.Error
}
