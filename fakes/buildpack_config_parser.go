package fakes

import "sync"

type BuildpackConfigParser struct {
	ParseProjectPathCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			ProjectPath string
			Err         error
		}
		Stub func(string) (string, error)
	}
}

func (f *BuildpackConfigParser) ParseProjectPath(param1 string) (string, error) {
	f.ParseProjectPathCall.Lock()
	defer f.ParseProjectPathCall.Unlock()
	f.ParseProjectPathCall.CallCount++
	f.ParseProjectPathCall.Receives.Path = param1
	if f.ParseProjectPathCall.Stub != nil {
		return f.ParseProjectPathCall.Stub(param1)
	}
	return f.ParseProjectPathCall.Returns.ProjectPath, f.ParseProjectPathCall.Returns.Err
}
