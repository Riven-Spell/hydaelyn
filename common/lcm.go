package common

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	LCMServiceNameLog           = "log"
	LCMServiceNameConfig        = "config"
	LCMServiceNameBot           = "bot"
	LCMServiceNameSQL           = "SQL"
	LCMServiceNameRoleReact     = "RoleReact"
	LCMServiceNameAutoScheduler = "AutoScheduler"
)

type LCMService struct {
	Name         string
	Dependencies []string
	GetSvc       func() interface{}
	Shutdown     func() error
	Startup      func(deps []interface{}) error
}

var singleLCM *LifeCycleManager

type LifeCycleManager struct {
	Services map[string]LCMService
}

func RunHookIfNotNil(hook func() error) error {
	if hook != nil {
		return hook()
	}

	return nil
}

func (lcm *LifeCycleManager) getServiceRing(sName string) uint {
	svc, ok := lcm.Services[sName]

	if !ok {
		return 0
	}

	// we start at ring 0. This means we have no dependencies.
	ring := uint(0)
	for _, v := range svc.Dependencies {
		if newRing := lcm.getServiceRing(v); newRing > ring {
			ring = newRing
		}
	}

	return ring
}

// Shutdown shuts down the lifecycle manager, if it exists.
func (lcm *LifeCycleManager) Shutdown() {
	if lcm == nil {
		return
	}

	errors := make(map[string]error)

	// Establish service "rings" and shutdown from the top-down.
	highestRing := uint(0)
	rings := make(map[uint][]string)

	for k := range lcm.Services {
		serviceRing := lcm.getServiceRing(k)
		rings[serviceRing] = append(rings[serviceRing], k)

		if serviceRing > highestRing {
			highestRing = serviceRing
		}
	}

	// Walk down the service rings.
	for {
		for _, v := range rings[highestRing] {
			err := RunHookIfNotNil(lcm.Services[v].Shutdown)

			if err != nil {
				errors[v] = err
			}
		}

		if highestRing == 0 {
			break
		}
		highestRing--
	}

	if lcm == singleLCM {
		singleLCM = nil // Kill off the LCM
	}
}

// GetLifeCycleManager returns the existing lifecycle manager, or, if none exists, creates a default.
func GetLifeCycleManager() *LifeCycleManager {
	if singleLCM == nil {
		return CreateLifeCycleManager(nil)
	}

	return singleLCM
}

// CreateLifeCycleManager creates a LCM and puts it into the singleton instance; accessible with GetLifeCycleManager()
func CreateLifeCycleManager(services []LCMService) *LifeCycleManager {
	// Initialize the LCM
	singleLCM = &LifeCycleManager{
		Services: make(map[string]LCMService),
	}

	// Register services
	for _, v := range services {
		singleLCM.Services[v.Name] = v
	}

	// Start services in order of dependency
	toStart := make(map[string]LCMService)
	for k, v := range singleLCM.Services {
		toStart[k] = v
	}
	started := make(map[string]bool)
	errored := make(map[string]error)

	for len(toStart) > 0 {
		for _, v := range toStart {
			// Can we start the service yet?
			startable := true
			failedDeps := make([]string, 0)
			for _, dep := range v.Dependencies {
				if _, ok := errored[dep]; ok { // If a dependency errored out, we definitely cannot start this service.
					failedDeps = append(failedDeps, dep)
					startable = false
				}

				if _, ok := started[dep]; !ok { // If a dependency hasn't started yet, we can't start this service yet.
					startable = false
				}
			}

			if !startable {
				if len(failedDeps) > 0 { // We cannot ever start this service, mark it as errored, and list which dependencies failed.
					str := "dependencies "
					for _, v := range failedDeps {
						str += v + ","
					}
					str = str[:len(str)-1] // trim last comma
					str += " failed to start"

					errored[v.Name] = errors.New(str)

					delete(toStart, v.Name) // we cannot ever start this; don't consider it anymore.
				}

				continue
			}

			var wrappedStartup func() error
			if v.Startup != nil {
				wrappedStartup = func() error {
					deps := make([]interface{}, len(v.Dependencies))

					for k, v := range v.Dependencies {
						deps[k] = singleLCM.Services[v].GetSvc()
					}

					return v.Startup(deps)
				}
			}

			err := RunHookIfNotNil(wrappedStartup)

			if err != nil {
				errored[v.Name] = err
			} else {
				started[v.Name] = true
			}

			delete(toStart, v.Name)
		}
	}

	if len(errored) > 0 {
		logErr := func(err error) {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}

		// Grab the logger, if it's available.
		if _, ok := started[LCMServiceNameLog]; ok {
			logger := singleLCM.Services[LCMServiceNameLog].GetSvc().(*log.Logger)

			logErr = func(err error) {
				logger.Println(err)
			}
		}

		for k, v := range errored {
			logErr(fmt.Errorf("service %s failed to start: %w", k, v))
		}

		os.Exit(1)
	}

	return singleLCM
}
