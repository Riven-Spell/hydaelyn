package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

// ========= LCM ==========

func LCMServiceLogger() LCMService {
	var lcmLog *log.Logger
	var swapLog *ReplaceableWriter
	var shutdownCh chan bool

	return LCMService{
		Name: LCMServiceNameLog,
		Startup: func() error {
			rootCfg, err := ReadConfig()
			if err != nil {
				return err
			}

			cfg := rootCfg.Log

			writers := []io.Writer{os.Stdout}
			if cfg.RetainLogDays > 0 {
				f, err := GetLog(cfg.LogDir)
				if err != nil {
					return err
				}

				swapLog = &ReplaceableWriter{&sync.Mutex{}, f}
				writers = append(writers, swapLog)
			}

			lcmLog = log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

			go func() {
				// Replace the log whenever the day goes by.
				await := time.After(time.Until(time.Now().Truncate(time.Hour * 24).Add(time.Hour * 24)))

				for {
					select {
					case <-shutdownCh:
						return // stop making new logs.
					case <-await:
						// make a new log.
						time.Sleep(time.Second) // make sure we're fully in the new day
						// generate a new log.
						orig := swapLog.w
						f, err := GetLog(cfg.LogDir)
						if err != nil {
							// not sure why this happened, but we'll ignore it and keep writing to the old log.
							continue
						}

						swapLog.ReplaceWriter(f)    // swap for our new log
						_ = orig.(*os.File).Close() // close the old log
						// todo: truncate old days.

						await = time.After(time.Until(time.Now().Truncate(time.Hour * 24).Add(time.Hour * 24)))
					}
				}
			}()

			return nil
		},
		GetSvc: func() interface{} {
			return lcmLog
		},
		Shutdown: func() error {
			return nil
		},
	}
}

// ========== CORE ==========

func GetLog(dir string) (*os.File, error) {
	now := time.Now()
	name := fmt.Sprintf("hydaelyn-%d-%d-%d.log", now.Year(), now.Month(), now.Day())

	os.MkdirAll(dir, 0770)

	return os.OpenFile(path.Join(dir, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
}
